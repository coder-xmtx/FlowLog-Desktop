// Package store 用 SQLite 保存流记的全部数据。
//
// 表结构（见 schema）：
//
//	days     一天一行，主键是 "2026-11-13"
//	entries  学习条目，属于某一天，按 position 排序
//	points   要点，属于某条 entry
//	todos    待办，属于某一天
//	settings 键值对，存外观与用户称呼
//
// 写操作走 sqlite.DB.Transaction：一天的数据整体替换，
// 要么全写成，要么全不写，不会留下半天的数据。
package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/plugins/sqlite"
)

// FileName 是数据库文件名。
const FileName = "flowlog.db"

// schema 是建表语句。用 IF NOT EXISTS，可以重复执行。
var schema = []string{
	`CREATE TABLE IF NOT EXISTS days (
		date       TEXT PRIMARY KEY,
		created_at INTEGER NOT NULL DEFAULT 0
	)`,
	`CREATE TABLE IF NOT EXISTS entries (
		id        TEXT PRIMARY KEY,
		date      TEXT NOT NULL REFERENCES days(date) ON DELETE CASCADE,
		subject   TEXT NOT NULL,
		chapter   TEXT NOT NULL DEFAULT '',
		section   TEXT NOT NULL DEFAULT '',
		note      TEXT NOT NULL DEFAULT '',
		minutes   INTEGER NOT NULL DEFAULT 0,
		at        TEXT NOT NULL DEFAULT '',
		position  INTEGER NOT NULL DEFAULT 0
	)`,
	`CREATE TABLE IF NOT EXISTS points (
		id       TEXT PRIMARY KEY,
		entry_id TEXT NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
		text     TEXT NOT NULL,
		done     INTEGER NOT NULL DEFAULT 0,
		position INTEGER NOT NULL DEFAULT 0
	)`,
	`CREATE TABLE IF NOT EXISTS todos (
		id       TEXT PRIMARY KEY,
		date     TEXT NOT NULL REFERENCES days(date) ON DELETE CASCADE,
		text     TEXT NOT NULL,
		done     INTEGER NOT NULL DEFAULT 0,
		at       TEXT NOT NULL DEFAULT '',
		position INTEGER NOT NULL DEFAULT 0
	)`,
	`CREATE TABLE IF NOT EXISTS settings (
		key   TEXT PRIMARY KEY,
		value TEXT NOT NULL
	)`,
	`CREATE INDEX IF NOT EXISTS entries_by_date ON entries(date, position)`,
	`CREATE INDEX IF NOT EXISTS points_by_entry ON points(entry_id, position)`,
	`CREATE INDEX IF NOT EXISTS todos_by_date ON todos(date, position)`,
}

// SchemaVersion 是当前表结构版本，记在 PRAGMA user_version 里。
const SchemaVersion = 1

// Store 是一个打开的数据库。
type Store struct {
	db   *sqlite.DB
	path string
}

// Open 打开（必要时创建）数据库，并保证表结构是最新的。
func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("创建数据目录: %w", err)
	}
	db, err := sqlite.Open(context.Background(), path, sqlite.OpenOptions{})
	if err != nil {
		return nil, fmt.Errorf("打开数据库 %s: %w", path, err)
	}
	s := &Store{db: db, path: path}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// migrate 建表并把 user_version 推到 SchemaVersion。
func (s *Store) migrate() error {
	ctx := context.Background()
	stmts := make([]sqlite.Statement, 0, len(schema))
	for _, sql := range schema {
		stmts = append(stmts, sqlite.Statement{SQL: sql})
	}
	if _, err := s.db.Transaction(ctx, stmts); err != nil {
		return fmt.Errorf("建表: %w", err)
	}
	if _, err := s.db.Execute(ctx, fmt.Sprintf("PRAGMA user_version = %d", SchemaVersion)); err != nil {
		return fmt.Errorf("写 schema 版本: %w", err)
	}
	return nil
}

// Close 关闭数据库。重复调用无副作用。
func (s *Store) Close() error { return s.db.Close() }

// Path 返回数据库文件的绝对路径。
func (s *Store) Path() string {
	if abs, err := filepath.Abs(s.path); err == nil {
		return abs
	}
	return s.path
}

// Size 返回数据库文件当前的字节数，用于设置页展示。
func (s *Store) Size() int64 {
	if info, err := os.Stat(s.path); err == nil {
		return info.Size()
	}
	return 0
}

// ---------------------------------------------------------------- 读

// Days 读出全部日子，按日期从新到旧，并带上各自的条目、要点与待办。
func (s *Store) Days() ([]model.Day, error) {
	ctx := context.Background()

	rows, err := s.db.Query(ctx, `SELECT date FROM days ORDER BY date DESC`)
	if err != nil {
		return nil, fmt.Errorf("读取日期: %w", err)
	}
	days := make([]model.Day, 0, len(rows.Values))
	index := make(map[string]int, len(rows.Values))
	for _, row := range rows.Values {
		date, _ := row[0].(string)
		if date == "" {
			continue
		}
		index[date] = len(days)
		days = append(days, model.Day{Date: date, Entries: []model.Entry{}, Todos: []model.Todo{}})
	}
	if len(days) == 0 {
		return days, nil
	}

	if err := s.loadEntries(days, index); err != nil {
		return nil, err
	}
	if err := s.loadTodos(days, index); err != nil {
		return nil, err
	}
	return days, nil
}

func (s *Store) loadEntries(days []model.Day, index map[string]int) error {
	ctx := context.Background()
	rows, err := s.db.Query(ctx, `SELECT id, date, subject, chapter, section, note, minutes, at
		FROM entries ORDER BY date, position`)
	if err != nil {
		return fmt.Errorf("读取条目: %w", err)
	}

	entryDay := make(map[string]int, len(rows.Values)) // entry id -> day index
	entryAt := make(map[string]int, len(rows.Values))  // entry id -> entry index
	for _, row := range rows.Values {
		id := text(row[0])
		date := text(row[1])
		di, ok := index[date]
		if !ok {
			continue
		}
		days[di].Entries = append(days[di].Entries, model.Entry{
			ID:      id,
			Subject: text(row[2]),
			Chapter: text(row[3]),
			Section: text(row[4]),
			Note:    text(row[5]),
			Minutes: int(number(row[6])),
			At:      text(row[7]),
			Points:  []model.Point{},
		})
		entryDay[id] = di
		entryAt[id] = len(days[di].Entries) - 1
	}

	points, err := s.db.Query(ctx, `SELECT entry_id, id, text, done FROM points ORDER BY entry_id, position`)
	if err != nil {
		return fmt.Errorf("读取要点: %w", err)
	}
	for _, row := range points.Values {
		entryID := text(row[0])
		di, ok := entryDay[entryID]
		if !ok {
			continue
		}
		ei := entryAt[entryID]
		days[di].Entries[ei].Points = append(days[di].Entries[ei].Points, model.Point{
			ID:   text(row[1]),
			Text: text(row[2]),
			Done: number(row[3]) != 0,
		})
	}
	return nil
}

func (s *Store) loadTodos(days []model.Day, index map[string]int) error {
	ctx := context.Background()
	rows, err := s.db.Query(ctx, `SELECT date, id, text, done, at FROM todos ORDER BY date, position`)
	if err != nil {
		return fmt.Errorf("读取待办: %w", err)
	}
	for _, row := range rows.Values {
		di, ok := index[text(row[0])]
		if !ok {
			continue
		}
		days[di].Todos = append(days[di].Todos, model.Todo{
			ID:   text(row[1]),
			Text: text(row[2]),
			Done: number(row[3]) != 0,
			At:   text(row[4]),
		})
	}
	return nil
}

// ---------------------------------------------------------------- 写

// SaveDay 用 day 的内容整体替换某一天。
//
// 一天没有条目也没有待办时，连这一天的行一起删掉：
// 界面上不会出现一张空卡片，数据库里也不留空壳。
func (s *Store) SaveDay(day model.Day) error {
	if err := validDate(day.Date); err != nil {
		return err
	}
	day.Normalize()

	ctx := context.Background()
	if !day.HasContent() {
		return s.DeleteDay(day.Date)
	}

	stmts := make([]sqlite.Statement, 0, 8+len(day.Entries)*2+len(day.Todos))
	stmts = append(stmts,
		sqlite.Statement{SQL: `INSERT OR IGNORE INTO days (date, created_at) VALUES (?, ?)`,
			Args: []any{day.Date, time.Now().Unix()}},
		sqlite.Statement{SQL: `DELETE FROM entries WHERE date = ?`, Args: []any{day.Date}},
		sqlite.Statement{SQL: `DELETE FROM todos WHERE date = ?`, Args: []any{day.Date}},
	)
	for i, e := range day.Entries {
		entryID := e.ID
		if entryID == "" {
			entryID = model.NewID()
		}
		stmts = append(stmts,
			sqlite.Statement{SQL: `INSERT INTO entries
				(id, date, subject, chapter, section, note, minutes, at, position)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				Args: []any{entryID, day.Date, e.Subject, e.Chapter, e.Section, e.Note, e.Minutes, e.At, i}},
			sqlite.Statement{SQL: `DELETE FROM points WHERE entry_id = ?`, Args: []any{entryID}},
		)
		for j, p := range e.Points {
			pointID := p.ID
			if pointID == "" {
				pointID = model.NewID()
			}
			stmts = append(stmts, sqlite.Statement{SQL: `INSERT INTO points
				(id, entry_id, text, done, position) VALUES (?, ?, ?, ?, ?)`,
				Args: []any{pointID, entryID, p.Text, boolInt(p.Done), j}})
		}
	}
	for i, t := range day.Todos {
		todoID := t.ID
		if todoID == "" {
			todoID = model.NewID()
		}
		stmts = append(stmts, sqlite.Statement{SQL: `INSERT INTO todos
			(id, date, text, done, at, position) VALUES (?, ?, ?, ?, ?, ?)`,
			Args: []any{todoID, day.Date, t.Text, boolInt(t.Done), t.At, i}})
	}
	if _, err := s.db.Transaction(ctx, stmts); err != nil {
		return fmt.Errorf("保存 %s: %w", day.Date, err)
	}
	return nil
}

// DeleteDay 删除一天和它名下的条目、要点、待办。
func (s *Store) DeleteDay(date string) error {
	if err := validDate(date); err != nil {
		return err
	}
	// entries 由外键级联删除，points 再由 entries 级联。
	if _, err := s.db.Execute(context.Background(), `DELETE FROM days WHERE date = ?`, date); err != nil {
		return fmt.Errorf("删除 %s: %w", date, err)
	}
	return nil
}

// DeleteAll 清空所有数据，用于「从备份恢复」前的一次性清理。
func (s *Store) DeleteAll() error {
	if _, err := s.db.Execute(context.Background(), `DELETE FROM days`); err != nil {
		return fmt.Errorf("清空数据: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------- 设置

// Settings 读出应用设置，缺失的键用默认值补齐。
func (s *Store) Settings() (model.Settings, error) {
	st := model.Settings{Theme: model.ThemeLight}
	rows, err := s.db.Query(context.Background(), `SELECT key, value FROM settings`)
	if err != nil {
		return st, fmt.Errorf("读取设置: %w", err)
	}
	for _, row := range rows.Values {
		switch text(row[0]) {
		case "theme":
			st.Theme = text(row[1])
		case "userName":
			st.UserName = text(row[1])
		case "onboarded":
			st.Onboarded = text(row[1]) == "1"
		}
	}
	if st.Theme != model.ThemeDark {
		st.Theme = model.ThemeLight
	}
	return st, nil
}

// SaveSettings 写入应用设置。
func (s *Store) SaveSettings(st model.Settings) error {
	if st.Theme != model.ThemeDark {
		st.Theme = model.ThemeLight
	}
	onboarded := "0"
	if st.Onboarded {
		onboarded = "1"
	}
	stmts := []sqlite.Statement{
		{SQL: `INSERT OR REPLACE INTO settings (key, value) VALUES ('theme', ?)`, Args: []any{st.Theme}},
		{SQL: `INSERT OR REPLACE INTO settings (key, value) VALUES ('userName', ?)`, Args: []any{st.UserName}},
		{SQL: `INSERT OR REPLACE INTO settings (key, value) VALUES ('onboarded', ?)`, Args: []any{onboarded}},
	}
	if _, err := s.db.Transaction(context.Background(), stmts); err != nil {
		return fmt.Errorf("保存设置: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------- 小工具

// validDate 校验日期格式，也避免拼出意外的路径或主键。
func validDate(date string) error {
	if _, err := model.ParseDate(date); err != nil {
		return fmt.Errorf("非法日期 %q：需要 2006-01-02 格式", date)
	}
	return nil
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// text 把查询结果安全地读成字符串。
func text(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case []byte:
		return string(x)
	}
	return ""
}

// number 把查询结果安全地读成整数（SQLite 的 INTEGER 是 int64）。
func number(v any) int64 {
	switch x := v.(type) {
	case int64:
		return x
	case float64:
		return int64(x)
	case string:
		var n int64
		fmt.Sscan(x, &n)
		return n
	}
	return 0
}

// SortDays 按日期倒序整理切片（导入时用）。
func SortDays(days []model.Day) {
	sort.Slice(days, func(i, j int) bool { return days[i].Date > days[j].Date })
}

// ErrNotBackup 表示文件不是流记的备份。
var ErrNotBackup = errors.New("不是流记的备份文件")
