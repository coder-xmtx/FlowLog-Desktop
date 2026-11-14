package store

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/plugins/sqlite"
)

func newStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), FileName))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func sampleDay() model.Day {
	return model.Day{
		Date: "2026-10-10",
		Entries: []model.Entry{{
			ID: "e1", Subject: "数据结构", Chapter: "第 6 章 图", Section: "第 1 节 图的基本概念",
			Points: []model.Point{
				{ID: "p1", Text: "图的定义"},
				{ID: "p2", Text: "连通分量", Done: true},
			},
			Note: "这章要画图", Minutes: 45, At: "21:30",
		}},
	}
}

func sampleTodo() model.Todo {
	return model.Todo{
		ID: "t1", Text: "复习 B 树", Created: "2026-10-10", At: "09:10",
	}
}

func TestSaveAndLoadDay(t *testing.T) {
	s := newStore(t)
	if err := s.SaveDay(sampleDay()); err != nil {
		t.Fatalf("SaveDay: %v", err)
	}

	days, err := s.Days()
	if err != nil {
		t.Fatalf("Days: %v", err)
	}
	if len(days) != 1 {
		t.Fatalf("读到 %d 天, 期望 1", len(days))
	}
	got := days[0]
	if got.Date != "2026-10-10" {
		t.Errorf("日期 = %q", got.Date)
	}
	if len(got.Entries) != 1 || got.Entries[0].Subject != "数据结构" {
		t.Fatalf("条目丢失: %+v", got.Entries)
	}
	e := got.Entries[0]
	if e.Chapter != "第 6 章 图" || e.Section != "第 1 节 图的基本概念" {
		t.Errorf("章节字段 = %q / %q", e.Chapter, e.Section)
	}
	if len(e.Points) != 2 || !e.Points[1].Done {
		t.Errorf("要点 = %+v", e.Points)
	}
	if e.Minutes != 45 || e.At != "21:30" {
		t.Errorf("时长/时间 = %d / %q", e.Minutes, e.At)
	}
}

func TestSaveIsReplaceNotAppend(t *testing.T) {
	s := newStore(t)
	day := sampleDay()
	if err := s.SaveDay(day); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveDay(day); err != nil {
		t.Fatal(err)
	}
	days, err := s.Days()
	if err != nil {
		t.Fatal(err)
	}
	if len(days[0].Entries) != 1 {
		t.Fatalf("重复保存后: %d 条记录", len(days[0].Entries))
	}
}

func TestEditReplacesPoints(t *testing.T) {
	s := newStore(t)
	day := sampleDay()
	if err := s.SaveDay(day); err != nil {
		t.Fatal(err)
	}
	e := &day.Entries[0]
	e.Points = []model.Point{{ID: "p2", Text: "连通分量（改）", Done: true}, {ID: "p3", Text: "邻接矩阵"}}
	if err := s.SaveDay(day); err != nil {
		t.Fatal(err)
	}
	days, _ := s.Days()
	got := days[0].Entries[0].Points
	if len(got) != 2 {
		t.Fatalf("要点 = %+v", got)
	}
	if got[0].Text != "连通分量（改）" || !got[0].Done {
		t.Errorf("第一条 = %+v", got[0])
	}
	if got[1].Text != "邻接矩阵" {
		t.Errorf("第二条 = %+v", got[1])
	}
}

func TestEmptyDayIsRemoved(t *testing.T) {
	s := newStore(t)
	if err := s.SaveDay(sampleDay()); err != nil {
		t.Fatal(err)
	}
	// 清空后保存：这一天的行应该消失。
	if err := s.SaveDay(model.Day{Date: "2026-10-10"}); err != nil {
		t.Fatal(err)
	}
	days, err := s.Days()
	if err != nil {
		t.Fatal(err)
	}
	if len(days) != 0 {
		t.Fatalf("清空后还剩 %d 天", len(days))
	}
}

func TestDeleteDayCascadesEntriesOnly(t *testing.T) {
	s := newStore(t)
	if err := s.SaveDay(sampleDay()); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveTodo(sampleTodo()); err != nil {
		t.Fatal(err)
	}
	// 删掉这一天的日记，待办必须原样保留：两者已经解耦。
	if err := s.DeleteDay("2026-10-10"); err != nil {
		t.Fatal(err)
	}
	days, _ := s.Days()
	if len(days) != 0 {
		t.Fatalf("删除后还有 %d 天", len(days))
	}
	todos, err := s.Todos()
	if err != nil {
		t.Fatal(err)
	}
	if len(todos) != 1 || todos[0].ID != "t1" {
		t.Fatalf("待办被日记的删除影响了: %+v", todos)
	}
	// 级联删掉的条目 ID 之后还能复用。
	d := model.Day{Date: "2026-10-11", Entries: []model.Entry{{ID: "e1", Subject: "x"}}}
	if err := s.SaveDay(d); err != nil {
		t.Fatalf("重用被删除的 ID: %v", err)
	}
}

func TestDaysSortedNewestFirst(t *testing.T) {
	s := newStore(t)
	for _, date := range []string{"2026-01-01", "2026-03-05", "2025-12-31"} {
		if err := s.SaveDay(model.Day{Date: date, Entries: []model.Entry{{ID: date, Subject: "x"}}}); err != nil {
			t.Fatal(err)
		}
	}
	days, err := s.Days()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"2026-03-05", "2026-01-01", "2025-12-31"}
	for i, date := range want {
		if days[i].Date != date {
			t.Errorf("第 %d 天 = %q, 期望 %q", i, days[i].Date, date)
		}
	}
}

func TestSaveTodoAndOrder(t *testing.T) {
	s := newStore(t)
	todos := []model.Todo{
		{ID: "a", Text: "没有截止", Created: "2026-11-10"},
		{ID: "b", Text: "后天截止", Created: "2026-11-10", Due: "2026-11-16"},
		{ID: "c", Text: "明天截止", Created: "2026-11-11", Due: "2026-11-15"},
		{ID: "d", Text: "已完成的", Created: "2026-11-12", Due: "2026-11-12", Done: true},
	}
	for _, todo := range todos {
		if err := s.SaveTodo(todo); err != nil {
			t.Fatalf("SaveTodo(%s): %v", todo.ID, err)
		}
	}
	got, err := s.Todos()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("读到 %d 条待办", len(got))
	}
	// 未完成在前：有截止日期按日期升序，没日期的最后；已完成压最后。
	wantIDs := []string{"c", "b", "a", "d"}
	for i, id := range wantIDs {
		if got[i].ID != id {
			t.Errorf("第 %d 条 = %q, 期望 %q（顺序 %v）", i, got[i].ID, id, ids(got))
		}
	}
	if got[0].Text != "明天截止" || got[0].Due != "2026-11-15" {
		t.Errorf("内容丢失: %+v", got[0])
	}
	if !got[3].Done {
		t.Errorf("完成状态丢失: %+v", got[3])
	}
}

func ids(todos []model.Todo) []string {
	out := make([]string, len(todos))
	for i, t := range todos {
		out[i] = t.ID
	}
	return out
}

func TestTodoUpdateAndDelete(t *testing.T) {
	s := newStore(t)
	todo := sampleTodo()
	if err := s.SaveTodo(todo); err != nil {
		t.Fatal(err)
	}
	todo.Done = true
	todo.Text = "复习 B 树（改）"
	if err := s.SaveTodo(todo); err != nil {
		t.Fatal(err)
	}
	got, _ := s.Todos()
	if len(got) != 1 || !got[0].Done || got[0].Text != "复习 B 树（改）" {
		t.Fatalf("更新失败: %+v", got)
	}
	if err := s.DeleteTodo(todo.ID); err != nil {
		t.Fatal(err)
	}
	got, _ = s.Todos()
	if len(got) != 0 {
		t.Fatalf("删除后还有 %d 条", len(got))
	}
}

func TestTodoValidation(t *testing.T) {
	s := newStore(t)
	cases := []model.Todo{
		{ID: "", Text: "没有 ID"},
		{ID: "x", Text: "   "},
		{ID: "y", Text: "日期不对", Due: "2026-13-40"},
	}
	for i, todo := range cases {
		if err := s.SaveTodo(todo); err == nil {
			t.Errorf("第 %d 条应该被拒绝: %+v", i, todo)
		}
	}
	// 没有截止日期是合法的。
	if err := s.SaveTodo(model.Todo{ID: "ok", Text: "无截止"}); err != nil {
		t.Errorf("没有截止日期应该可保存: %v", err)
	}
}

func TestRejectBadDates(t *testing.T) {
	s := newStore(t)
	for _, date := range []string{"", "2026-13-01", "../evil", "2026-1-1", "20261113"} {
		if err := s.SaveDay(model.Day{Date: date, Entries: []model.Entry{{ID: "1", Subject: "x"}}}); err == nil {
			t.Errorf("SaveDay(%q) 应该报错", date)
		}
		if err := s.DeleteDay(date); err == nil {
			t.Errorf("DeleteDay(%q) 应该报错", date)
		}
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	s := newStore(t)
	st, err := s.Settings()
	if err != nil {
		t.Fatalf("Settings: %v", err)
	}
	if st.Theme != model.ThemeDark {
		t.Errorf("默认外观 = %q, 期望 dark", st.Theme)
	}
	if err := s.SaveSettings(model.Settings{Theme: model.ThemeSystem, UserName: "Mixsu", Onboarded: true}); err != nil {
		t.Fatal(err)
	}
	got, err := s.Settings()
	if err != nil {
		t.Fatal(err)
	}
	if got.Theme != model.ThemeSystem || got.UserName != "Mixsu" || !got.Onboarded {
		t.Errorf("设置往返 = %+v", got)
	}
	// 未知主题收敛为默认（深色）
	if err := s.SaveSettings(model.Settings{Theme: "neon"}); err != nil {
		t.Fatal(err)
	}
	got, _ = s.Settings()
	if got.Theme != model.ThemeDark {
		t.Errorf("未知主题 = %q, 期望被收敛成 dark", got.Theme)
	}
}

func TestBackupRoundTrip(t *testing.T) {
	s := newStore(t)
	if err := s.SaveDay(sampleDay()); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveTodo(model.Todo{ID: "t9", Text: "带截止的待办", Created: "2026-10-10", Due: "2026-10-12"}); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveSettings(model.Settings{Theme: model.ThemeSystem, UserName: "Mixsu", Onboarded: true}); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(t.TempDir(), "backup.json")
	if err := s.ExportFile(path); err != nil {
		t.Fatalf("导出: %v", err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"FlowLog"`, `"数据结构"`, `"第 6 章 图"`, `"Mixsu"`, `"t9"`, `"2026-10-12"`} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("备份里缺少 %s:\n%s", want, raw)
		}
	}

	b, err := ReadBackup(path)
	if err != nil {
		t.Fatalf("ReadBackup: %v", err)
	}
	other, err := Open(filepath.Join(t.TempDir(), FileName))
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	if err := other.Restore(b); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	days, _ := other.Days()
	if len(days) != 1 || days[0].PointCount() != 2 {
		t.Fatalf("恢复后 = %+v", days)
	}
	todos, _ := other.Todos()
	if len(todos) != 1 || todos[0].Due != "2026-10-12" {
		t.Fatalf("待办没恢复: %+v", todos)
	}
	st, _ := other.Settings()
	if st.UserName != "Mixsu" || st.Theme != model.ThemeSystem {
		t.Errorf("设置没恢复: %+v", st)
	}
}

func TestRestoreReplacesEverything(t *testing.T) {
	s := newStore(t)
	if err := s.SaveDay(model.Day{Date: "2026-01-01", Entries: []model.Entry{{ID: "old", Subject: "旧数据"}}}); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveTodo(model.Todo{ID: "old-todo", Text: "旧待办"}); err != nil {
		t.Fatal(err)
	}
	b := &Backup{
		App: "FlowLog", Format: 2,
		Days:  []model.Day{{Date: "2026-02-02", Entries: []model.Entry{{ID: "new", Subject: "新数据"}}}},
		Todos: []model.Todo{{ID: "new-todo", Text: "新待办"}},
	}
	if err := s.Restore(b); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	days, _ := s.Days()
	if len(days) != 1 || days[0].Date != "2026-02-02" {
		t.Fatalf("恢复应该是替换而不是合并: %+v", days)
	}
	todos, _ := s.Todos()
	if len(todos) != 1 || todos[0].ID != "new-todo" {
		t.Fatalf("待办没被替换: %+v", todos)
	}
}

func TestReadBackupRejectsForeignFile(t *testing.T) {
	dir := t.TempDir()
	cases := map[string]string{
		"not-json.json":  "hello",
		"other-app.json": `{"app":"Other","format":2,"days":[]}`,
	}
	for name, body := range cases {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		if _, err := ReadBackup(path); !IsNotBackup(err) {
			t.Errorf("%s: 期望 ErrNotBackup, 得到 %v", name, err)
		}
	}
}

func TestDBFileIsCreated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", FileName)
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("数据库文件不存在: %v", err)
	}
	if !strings.HasSuffix(s.Path(), FileName) {
		t.Errorf("Path = %q", s.Path())
	}
}

// TestExportIsHumanReadable 确认导出的 JSON 不转义中文。
func TestExportIsHumanReadable(t *testing.T) {
	s := newStore(t)
	if err := s.SaveDay(sampleDay()); err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := s.Export(&buf); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "第 6 章 图") {
		t.Errorf("导出的 JSON 中文应原样可读:\n%s", buf.String())
	}
}

// TestMigrateTodosFromV1 确认 v0.3 的数据库能升到 v0.4：
// todos 的 date 变成 created，并补上空的 due。
func TestMigrateTodosFromV1(t *testing.T) {
	path := filepath.Join(t.TempDir(), FileName)

	// 手工造一个 v1 结构的库（todos 带 date 与外键）。
	db, err := sqlite.Open(context.Background(), path, sqlite.OpenOptions{})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	stmts := []sqlite.Statement{
		{SQL: `CREATE TABLE days (date TEXT PRIMARY KEY, created_at INTEGER NOT NULL DEFAULT 0)`},
		{SQL: `CREATE TABLE entries (id TEXT PRIMARY KEY, date TEXT NOT NULL REFERENCES days(date) ON DELETE CASCADE,
			subject TEXT NOT NULL, chapter TEXT NOT NULL DEFAULT '', section TEXT NOT NULL DEFAULT '',
			note TEXT NOT NULL DEFAULT '', minutes INTEGER NOT NULL DEFAULT 0, at TEXT NOT NULL DEFAULT '',
			position INTEGER NOT NULL DEFAULT 0)`},
		{SQL: `CREATE TABLE points (id TEXT PRIMARY KEY, entry_id TEXT NOT NULL REFERENCES entries(id) ON DELETE CASCADE,
			text TEXT NOT NULL, done INTEGER NOT NULL DEFAULT 0, position INTEGER NOT NULL DEFAULT 0)`},
		{SQL: `CREATE TABLE todos (id TEXT PRIMARY KEY, date TEXT NOT NULL REFERENCES days(date) ON DELETE CASCADE,
			text TEXT NOT NULL, done INTEGER NOT NULL DEFAULT 0, at TEXT NOT NULL DEFAULT '',
			position INTEGER NOT NULL DEFAULT 0)`},
		{SQL: `CREATE TABLE settings (key TEXT PRIMARY KEY, value TEXT NOT NULL)`},
		{SQL: `INSERT INTO days (date) VALUES ('2026-11-13')`},
		{SQL: `INSERT INTO todos (id, date, text, done, at, position) VALUES ('old1', '2026-11-13', '旧待办', 1, '09:10', 0)`},
		{SQL: `INSERT INTO settings (key, value) VALUES ('theme', 'light')`},
		{SQL: `PRAGMA user_version = 1`},
	}
	if _, err := db.Transaction(ctx, stmts); err != nil {
		t.Fatalf("造 v1 库: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open 升级: %v", err)
	}
	defer s.Close()
	todos, err := s.Todos()
	if err != nil {
		t.Fatal(err)
	}
	if len(todos) != 1 {
		t.Fatalf("升级后待办 = %+v", todos)
	}
	if todos[0].ID != "old1" || todos[0].Created != "2026-11-13" || todos[0].Due != "" || !todos[0].Done {
		t.Errorf("升级结果 = %+v", todos[0])
	}
	// v1 里那天的日记行还在。
	days, _ := s.Days()
	if len(days) != 1 || days[0].Date != "2026-11-13" {
		t.Errorf("升级后日期 = %+v", days)
	}
	if ver, _ := s.userVersion(); ver != SchemaVersion {
		t.Errorf("user_version = %d, 期望 %d", ver, SchemaVersion)
	}
	// 升级后的待办可以正常增删。
	if err := s.SaveTodo(model.Todo{ID: "new1", Text: "新待办", Due: "2026-11-20"}); err != nil {
		t.Errorf("升级后保存待办: %v", err)
	}
}
