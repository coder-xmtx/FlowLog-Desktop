package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"flowlog-desktop/internal/model"
)

func newStore(t *testing.T) (*Store, string) {
	t.Helper()
	dir := t.TempDir()
	s, err := New(dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s, dir
}

func TestSaveAndLoadDay(t *testing.T) {
	s, dir := newStore(t)
	day := model.Day{
		Date: "2026-10-10",
		Entries: []model.Entry{{
			ID: "e1", Subject: "数据结构", Topic: "图 >> 图的基本概念",
			Points: []model.Point{{ID: "p1", Text: "图的定义"}, {ID: "p2", Text: "连通分量", Done: true}},
			Note:   "这章要画图", Minutes: 45, At: "21:30",
		}},
		Todos: []model.Todo{{ID: "t1", Text: "复习 B 树", Done: true}},
	}
	if err := s.SaveDay(day); err != nil {
		t.Fatalf("SaveDay: %v", err)
	}

	// 文件确实写到了 data/<date>.json，而且是可读的 JSON。
	raw, err := os.ReadFile(filepath.Join(dir, "2026-10-10.json"))
	if err != nil {
		t.Fatalf("读文件: %v", err)
	}
	var onDisk model.Day
	if err := json.Unmarshal(raw, &onDisk); err != nil {
		t.Fatalf("文件不是合法 JSON: %v", err)
	}
	// 人读的 JSON 不该把 ">>" 转义成 \u003e\u003e。
	if !strings.Contains(string(raw), "图 >> 图的基本概念") {
		t.Errorf("章节路径应原样保留: %s", raw)
	}

	got, err := s.LoadDay("2026-10-10")
	if err != nil {
		t.Fatalf("LoadDay: %v", err)
	}
	if len(got.Entries) != 1 || got.Entries[0].Subject != "数据结构" {
		t.Fatalf("Entries 丢失: %+v", got.Entries)
	}
	if got.PointCount() != 2 || got.DonePoints() != 1 {
		t.Fatalf("要点统计不对: total=%d done=%d", got.PointCount(), got.DonePoints())
	}
	if done, total := got.TodoProgress(); done != 1 || total != 1 {
		t.Fatalf("待办统计不对: %d/%d", done, total)
	}
}

func TestMissingDayIsEmpty(t *testing.T) {
	s, _ := newStore(t)
	day, err := s.LoadDay("2026-01-02")
	if err != nil {
		t.Fatalf("LoadDay: %v", err)
	}
	if day.Date != "2026-01-02" {
		t.Errorf("Date = %q", day.Date)
	}
	if day.HasContent() {
		t.Errorf("缺失的文件应该读成空 Day")
	}
	// 空 Day 的切片也应该是非 nil 的，免得 UI 里到处判空。
	if day.Entries == nil || day.Todos == nil {
		t.Errorf("切片应为空切片而非 nil: %+v", day)
	}
}

func TestSaveEmptyDayRemovesFile(t *testing.T) {
	s, dir := newStore(t)
	if err := s.SaveDay(model.Day{Date: "2026-01-03", Todos: []model.Todo{{ID: "1", Text: "x"}}}); err != nil {
		t.Fatalf("SaveDay: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "2026-01-03.json")); err != nil {
		t.Fatalf("文件应该存在: %v", err)
	}
	// 内容被清空之后，文件应当消失。
	if err := s.SaveDay(model.Day{Date: "2026-01-03"}); err != nil {
		t.Fatalf("SaveDay empty: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "2026-01-03.json")); !os.IsNotExist(err) {
		t.Fatalf("空日记的文件应该被删掉, stat err = %v", err)
	}
}

func TestSaveOverwritesAtomically(t *testing.T) {
	s, dir := newStore(t)
	for i := 0; i < 3; i++ {
		day := model.Day{Date: "2026-02-02", Todos: []model.Todo{{ID: "1", Text: "第" + string(rune('a'+i)) + "次"}}}
		if err := s.SaveDay(day); err != nil {
			t.Fatalf("SaveDay #%d: %v", i, err)
		}
	}
	// 临时文件不该留在数据目录里。
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.Name() != "2026-02-02.json" {
			t.Errorf("数据目录里有多余文件 %q", e.Name())
		}
	}
}

func TestLoadDaysSortedAndSkipsForeignFiles(t *testing.T) {
	s, dir := newStore(t)
	for _, date := range []string{"2026-01-01", "2026-03-05", "2025-12-31"} {
		if err := s.SaveDay(model.Day{Date: date, Todos: []model.Todo{{ID: "1", Text: date}}}); err != nil {
			t.Fatal(err)
		}
	}
	// 不该被当成日期的文件
	if err := os.WriteFile(filepath.Join(dir, "settings.json"), []byte(`{"theme":"dark"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "notes.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatal(err)
	}
	// 损坏的日期文件：跳过，并报告出来。
	if err := os.WriteFile(filepath.Join(dir, "2026-02-02.json"), []byte(`{oops`), 0o644); err != nil {
		t.Fatal(err)
	}

	days, bad, err := s.LoadDays()
	if err != nil {
		t.Fatalf("LoadDays: %v", err)
	}
	if len(days) != 3 {
		t.Fatalf("读到 %d 天, 期望 3: %+v", len(days), days)
	}
	want := []string{"2026-03-05", "2026-01-01", "2025-12-31"}
	for i, date := range want {
		if days[i].Date != date {
			t.Errorf("第 %d 天 = %q, 期望 %q（应按日期倒序）", i, days[i].Date, date)
		}
	}
	if len(bad) != 1 || bad[0] != "2026-02-02" {
		t.Errorf("坏文件 = %v, 期望 [2026-02-02]", bad)
	}
}

func TestLegacyDiaryMigrates(t *testing.T) {
	s, dir := newStore(t)
	legacy := `{"date":"2026-11-13","diary":"今天开始写 FlowLog。","todos":[]}`
	if err := os.WriteFile(filepath.Join(dir, "2026-11-13.json"), []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	day, err := s.LoadDay("2026-11-13")
	if err != nil {
		t.Fatalf("LoadDay: %v", err)
	}
	if len(day.Entries) != 1 {
		t.Fatalf("旧日记没有迁移成 Entry: %+v", day)
	}
	if day.Entries[0].Note != "今天开始写 FlowLog。" {
		t.Errorf("迁移后内容 = %q", day.Entries[0].Note)
	}
	if day.LegacyDiary != "" {
		t.Errorf("迁移后不应再保留 diary 字段")
	}
}

func TestRejectBadDates(t *testing.T) {
	s, _ := newStore(t)
	for _, date := range []string{"", "2026-13-01", "../evil", "2026-1-1", "20261113"} {
		if err := s.SaveDay(model.Day{Date: date, Todos: []model.Todo{{ID: "1", Text: "x"}}}); err == nil {
			t.Errorf("SaveDay(%q) 应该报错", date)
		}
		if _, err := s.LoadDay(date); err == nil {
			t.Errorf("LoadDay(%q) 应该报错", date)
		}
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	s, _ := newStore(t)
	st, err := s.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings: %v", err)
	}
	if st.Theme != model.ThemeSystem {
		t.Errorf("默认外观 = %q, 期望 system", st.Theme)
	}
	if err := s.SaveSettings(model.Settings{Theme: model.ThemeDark, Accent: "#ea580c"}); err != nil {
		t.Fatalf("SaveSettings: %v", err)
	}
	got, err := s.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings: %v", err)
	}
	if got.Theme != model.ThemeDark || got.Accent != "#ea580c" {
		t.Errorf("设置往返 = %+v", got)
	}
	// 写入未知的 Theme 值会被拉回 system，避免 UI 拿到无法识别的值。
	if err := s.SaveSettings(model.Settings{Theme: "neon"}); err != nil {
		t.Fatal(err)
	}
	got, _ = s.LoadSettings()
	if got.Theme != model.ThemeSystem {
		t.Errorf("未知外观 = %q, 期望被收敛成 system", got.Theme)
	}
}
