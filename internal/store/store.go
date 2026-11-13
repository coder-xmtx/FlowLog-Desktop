// Package store 负责把 model.Day 落到 data/ 目录下的 JSON 文件。
//
// 设计要点：
//   - 一天一个文件，文件名就是日期，便于人肉查看、grep、Git 同步。
//   - 写入是原子的：先写同目录下的临时文件，再 rename 覆盖，
//     这样即使写到一半断电/崩溃，也不会得到半个 JSON。
//   - 读取时做兼容迁移：早期版本的 "diary" 字符串会变成一条 Entry。
//   - 空的一天（没有条目也没有待办）会删除对应文件，避免目录里堆空壳。
package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"flowlog-desktop/internal/model"
)

// daysDirName / settingsName 是数据目录里的文件约定。
const (
	settingsName = "settings.json"
	tempPattern  = ".tmp-*.json"
)

// Store 是一个数据目录。
type Store struct {
	dir string
}

// New 创建 Store，并确保数据目录存在。
func New(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("创建数据目录 %s: %w", dir, err)
	}
	return &Store{dir: dir}, nil
}

// Dir 返回数据目录。
func (s *Store) Dir() string { return s.dir }

// filePath 返回某一天的数据文件路径。
func (s *Store) filePath(date string) string {
	return filepath.Join(s.dir, date+".json")
}

// validDate 校验日期字符串，既保证格式正确，也避免把 "../x" 之类的路径拼进来。
func validDate(date string) error {
	if _, err := model.ParseDate(date); err != nil {
		return fmt.Errorf("非法日期 %q: 需要 2006-01-02 格式", date)
	}
	return nil
}

// LoadDay 读取某一天。文件不存在时返回一个空 Day（不是错误）。
func (s *Store) LoadDay(date string) (model.Day, error) {
	if err := validDate(date); err != nil {
		return model.Day{}, err
	}
	raw, err := os.ReadFile(s.filePath(date))
	if errors.Is(err, os.ErrNotExist) {
		return model.Day{Date: date, Entries: []model.Entry{}, Todos: []model.Todo{}}, nil
	}
	if err != nil {
		return model.Day{}, fmt.Errorf("读取 %s: %w", date, err)
	}
	var day model.Day
	if err := json.Unmarshal(raw, &day); err != nil {
		return model.Day{}, fmt.Errorf("解析 %s.json: %w", date, err)
	}
	// 文件名是日期的权威来源，防止文件里写错日期导致列表错乱。
	day.Date = date
	day = migrate(day)
	day.Normalize()
	return day, nil
}

// migrate 把早期版本的裸日记字符串转成一条 Entry。
func migrate(day model.Day) model.Day {
	if day.LegacyDiary == "" || len(day.Entries) > 0 {
		return day
	}
	text := strings.TrimSpace(day.LegacyDiary)
	if text == "" {
		return day
	}
	entry := model.Entry{
		ID:      model.NewID(),
		Subject: "随手记",
		Note:    text,
		At:      "00:00",
	}
	day.Entries = append(day.Entries, entry)
	day.LegacyDiary = ""
	return day
}

// encodeJSON 把值写成缩进的 JSON，并且不转义 <、>、&。
//
// 默认的 json.Marshal 会把 ">>" 写成 "\u003e\u003e"，对机器没差别，
// 但这份文件是给人看、给 Git 比的：章节路径里的 ">>" 应该原样出现。
func encodeJSON(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// SaveDay 原子地写入某一天。
func (s *Store) SaveDay(day model.Day) error {
	if err := validDate(day.Date); err != nil {
		return err
	}
	day.Normalize()

	// 空的一天不留文件。
	if !day.HasContent() {
		return s.DeleteDay(day.Date)
	}

	raw, err := encodeJSON(day)
	if err != nil {
		return fmt.Errorf("序列化 %s: %w", day.Date, err)
	}

	tmp, err := os.CreateTemp(s.dir, tempPattern)
	if err != nil {
		return fmt.Errorf("创建临时文件: %w", err)
	}
	tmpName := tmp.Name()
	// 失败路径统一清掉临时文件，成功时 rename 之后它已经不存在。
	defer os.Remove(tmpName)

	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return fmt.Errorf("写入临时文件: %w", err)
	}
	// 先落盘再改名：rename 之后还能看到完整内容。
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("同步临时文件: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("关闭临时文件: %w", err)
	}
	if err := os.Rename(tmpName, s.filePath(day.Date)); err != nil {
		return fmt.Errorf("替换 %s.json: %w", day.Date, err)
	}
	return nil
}

// DeleteDay 删除某一天的数据文件（不存在不算错误）。
func (s *Store) DeleteDay(date string) error {
	if err := validDate(date); err != nil {
		return err
	}
	if err := os.Remove(s.filePath(date)); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("删除 %s.json: %w", date, err)
	}
	return nil
}

// LoadDays 读取目录里所有合法的日期文件，按日期从新到旧排序。
//
// bad 是读不出来的文件名（JSON 损坏、编码不对等），调用方可以提示用户，
// 但不要因此让整个应用起不来：能读多少是多少。
func (s *Store) LoadDays() (days []model.Day, bad []string, err error) {
	names, err := s.dayFiles()
	if err != nil {
		return nil, nil, err
	}
	days = make([]model.Day, 0, len(names))
	for _, name := range names {
		day, err := s.LoadDay(name)
		if err != nil {
			bad = append(bad, name)
			continue
		}
		days = append(days, day)
	}
	sort.Slice(days, func(i, j int) bool { return days[i].Date > days[j].Date })
	return days, bad, nil
}

// dayFiles 返回目录里所有日期文件的日期（不含 .json），从新到旧。
func (s *Store) dayFiles() ([]string, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, fmt.Errorf("读取数据目录: %w", err)
	}
	dates := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		date := strings.TrimSuffix(name, ".json")
		if validDate(date) != nil {
			continue // settings.json 之类，跳过
		}
		dates = append(dates, date)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))
	return dates, nil
}

// LoadSettings 读取应用设置。文件不存在时返回默认值。
func (s *Store) LoadSettings() (model.Settings, error) {
	raw, err := os.ReadFile(filepath.Join(s.dir, settingsName))
	if errors.Is(err, os.ErrNotExist) {
		return defaultSettings(), nil
	}
	if err != nil {
		return defaultSettings(), fmt.Errorf("读取设置: %w", err)
	}
	var st model.Settings
	if err := json.Unmarshal(raw, &st); err != nil {
		// 设置损坏不应该拦住用户，回到默认值即可。
		return defaultSettings(), nil
	}
	return withDefaults(st), nil
}

// SaveSettings 原子地写入应用设置。
func (s *Store) SaveSettings(st model.Settings) error {
	st = withDefaults(st)
	raw, err := encodeJSON(st)
	if err != nil {
		return fmt.Errorf("序列化设置: %w", err)
	}

	tmp, err := os.CreateTemp(s.dir, tempPattern)
	if err != nil {
		return fmt.Errorf("创建临时文件: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(raw); err != nil {
		tmp.Close()
		return fmt.Errorf("写入设置: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("同步设置: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("关闭设置: %w", err)
	}
	if err := os.Rename(tmpName, filepath.Join(s.dir, settingsName)); err != nil {
		return fmt.Errorf("替换设置文件: %w", err)
	}
	return nil
}

func defaultSettings() model.Settings {
	return model.Settings{Theme: model.ThemeSystem}
}

// withDefaults 补齐空字段，保证 Theme 一定是一个已知值。
func withDefaults(st model.Settings) model.Settings {
	switch st.Theme {
	case model.ThemeLight, model.ThemeDark, model.ThemeSystem:
	default:
		st.Theme = model.ThemeSystem
	}
	return st
}
