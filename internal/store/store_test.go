package store

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"flowlog-desktop/internal/model"
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
		Todos: []model.Todo{
			{ID: "t1", Text: "复习 B 树", Done: true, At: "09:10"},
			{ID: "t2", Text: "做题", At: "09:12"},
		},
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
	if done, total := got.TodoProgress(); done != 1 || total != 2 {
		t.Errorf("待办进度 = %d/%d", done, total)
	}
	// 顺序要保留
	if got.Todos[0].ID != "t1" || got.Todos[1].ID != "t2" {
		t.Errorf("待办顺序 = %+v", got.Todos)
	}
}

func TestSaveIsReplaceNotAppend(t *testing.T) {
	s := newStore(t)
	day := sampleDay()
	if err := s.SaveDay(day); err != nil {
		t.Fatal(err)
	}
	// 同一份数据保存两次，不应该出现重复行。
	if err := s.SaveDay(day); err != nil {
		t.Fatal(err)
	}
	days, err := s.Days()
	if err != nil {
		t.Fatal(err)
	}
	if len(days[0].Entries) != 1 || len(days[0].Todos) != 2 {
		t.Fatalf("重复保存后: %d 条记录 / %d 条待办", len(days[0].Entries), len(days[0].Todos))
	}
}

func TestEditReplacesPoints(t *testing.T) {
	s := newStore(t)
	day := sampleDay()
	if err := s.SaveDay(day); err != nil {
		t.Fatal(err)
	}
	// 编辑这条记录：删掉一个要点，改一个，新增一个。
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
	day := sampleDay()
	if err := s.SaveDay(day); err != nil {
		t.Fatal(err)
	}
	// 清空后保存：这一天的行应该消失。
	if err := s.SaveDay(model.Day{Date: day.Date}); err != nil {
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

func TestDeleteDayCascades(t *testing.T) {
	s := newStore(t)
	if err := s.SaveDay(sampleDay()); err != nil {
		t.Fatal(err)
	}
	if err := s.DeleteDay("2026-10-10"); err != nil {
		t.Fatal(err)
	}
	days, _ := s.Days()
	if len(days) != 0 {
		t.Fatalf("删除后还有 %d 天", len(days))
	}
	// 级联删掉的条目、要点不应该残留：再存一天，ID 还能复用（主键没被占用）。
	d := model.Day{Date: "2026-10-11", Entries: []model.Entry{{ID: "e1", Subject: "x"}}}
	if err := s.SaveDay(d); err != nil {
		t.Fatalf("重用被删除的 ID: %v", err)
	}
}

func TestDaysSortedNewestFirst(t *testing.T) {
	s := newStore(t)
	for _, date := range []string{"2026-01-01", "2026-03-05", "2025-12-31"} {
		if err := s.SaveDay(model.Day{Date: date, Todos: []model.Todo{{ID: date, Text: date}}}); err != nil {
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

func TestRejectBadDates(t *testing.T) {
	s := newStore(t)
	for _, date := range []string{"", "2026-13-01", "../evil", "2026-1-1", "20261113"} {
		if err := s.SaveDay(model.Day{Date: date, Todos: []model.Todo{{ID: "1", Text: "x"}}}); err == nil {
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
	if st.Theme != model.ThemeLight {
		t.Errorf("默认外观 = %q, 期望 light", st.Theme)
	}
	if err := s.SaveSettings(model.Settings{Theme: model.ThemeDark, UserName: "Mixsu", Onboarded: true}); err != nil {
		t.Fatal(err)
	}
	got, err := s.Settings()
	if err != nil {
		t.Fatal(err)
	}
	if got.Theme != model.ThemeDark || got.UserName != "Mixsu" || !got.Onboarded {
		t.Errorf("设置往返 = %+v", got)
	}
	// 未知主题收敛为 light
	if err := s.SaveSettings(model.Settings{Theme: "neon"}); err != nil {
		t.Fatal(err)
	}
	got, _ = s.Settings()
	if got.Theme != model.ThemeLight {
		t.Errorf("未知主题 = %q, 期望被收敛成 light", got.Theme)
	}
}

func TestBackupRoundTrip(t *testing.T) {
	s := newStore(t)
	if err := s.SaveDay(sampleDay()); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveSettings(model.Settings{Theme: model.ThemeDark, UserName: "Mixsu", Onboarded: true}); err != nil {
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
	for _, want := range []string{`"FlowLog"`, `"数据结构"`, `"第 6 章 图"`, `"Mixsu"`} {
		if !strings.Contains(string(raw), want) {
			t.Errorf("备份里缺少 %s:\n%s", want, raw)
		}
	}

	// 导进一个空库，内容应完全一致。
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
	days, err := other.Days()
	if err != nil {
		t.Fatal(err)
	}
	if len(days) != 1 || days[0].PointCount() != 2 || len(days[0].Todos) != 2 {
		t.Fatalf("恢复后 = %+v", days)
	}
	st, _ := other.Settings()
	if st.UserName != "Mixsu" || st.Theme != model.ThemeDark {
		t.Errorf("设置没恢复: %+v", st)
	}
}

func TestRestoreReplacesEverything(t *testing.T) {
	s := newStore(t)
	if err := s.SaveDay(model.Day{Date: "2026-01-01", Todos: []model.Todo{{ID: "old", Text: "旧数据"}}}); err != nil {
		t.Fatal(err)
	}
	b := &Backup{App: "FlowLog", Format: 1, Days: []model.Day{{Date: "2026-02-02", Todos: []model.Todo{{ID: "new", Text: "新数据"}}}}}
	if err := s.Restore(b); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	days, _ := s.Days()
	if len(days) != 1 || days[0].Date != "2026-02-02" {
		t.Fatalf("恢复应该是替换而不是合并: %+v", days)
	}
}

func TestReadBackupRejectsForeignFile(t *testing.T) {
	dir := t.TempDir()
	cases := map[string]string{
		"not-json.json":  "hello",
		"other-app.json": `{"app":"Other","format":1,"days":[]}`,
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

func TestImportJSONDir(t *testing.T) {
	dir := t.TempDir()
	// 老格式：v0.2 的 topic、v0.1 的 diary，顺带一个坏文件和一个 settings.json。
	files := map[string]string{
		"2026-11-13.json": `{"date":"2026-11-13","entries":[
			{"id":"e9","subject":"随手记","note":"今天开始写 FlowLog。","at":"00:00"},
			{"id":"e8","subject":"计算机组成原理","topic":"第 4 章 存储系统 >> 第 8 节 虚拟存储器",
			 "points":[{"id":"p1","text":"地址映射"}]}],
			"todos":[{"id":"1","text":"跑通 store","done":true}]}`,
		"2026-11-12.json": `{"date":"2026-11-12","todos":[{"id":"2","text":"第二天的待办"}]}`,
		"2026-11-10.json": `{"date":"2026-11-10","diary":"只有一段文字"}`,
		"settings.json":   `{"theme":"dark"}`,
		"2026-11-11.json": "not json",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	days, bad, err := ImportJSONDir(dir)
	if err != nil {
		t.Fatalf("ImportJSONDir: %v", err)
	}
	if len(days) != 3 {
		t.Fatalf("读到 %d 天: %+v", len(days), days)
	}
	if len(bad) != 1 || bad[0] != "2026-11-11.json" {
		t.Errorf("坏文件 = %v", bad)
	}
	if days[0].Date != "2026-11-13" {
		t.Errorf("应按日期倒序: %+v", days)
	}

	// v0.2 的 topic 应该被拆成章 / 节。
	var migrated model.Entry
	for _, e := range days[0].Entries {
		if e.Subject == "计算机组成原理" {
			migrated = e
		}
	}
	if migrated.Chapter != "第 4 章 存储系统" || migrated.Section != "第 8 节 虚拟存储器" {
		t.Errorf("topic 没有迁移成章 / 节: %+v", migrated)
	}
	if len(migrated.Points) != 1 || migrated.Points[0].Text != "地址映射" {
		t.Errorf("要点丢失: %+v", migrated.Points)
	}
	if done, _ := days[0].TodoProgress(); done != 1 {
		t.Errorf("待办的完成状态丢失: %+v", days[0].Todos)
	}

	// v0.1 的 diary 应该变成一条「随手记」。
	if len(days[2].Entries) != 1 || days[2].Entries[0].Subject != "随手记" {
		t.Errorf("diary 没有迁移: %+v", days[2].Entries)
	}

	s := newStore(t)
	n, err := s.ImportDays(days)
	if err != nil || n != 3 {
		t.Fatalf("ImportDays = %d, %v", n, err)
	}
	got, _ := s.Days()
	if len(got) != 3 || got[1].Todos[0].Text != "第二天的待办" {
		t.Errorf("导入结果 = %+v", got)
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

// TestExportIsHumanReadable 确认导出的 JSON 不转义 ">>"。
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
