package view

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"flowlog-desktop/internal/model"
	"flowlog-desktop/internal/store"

	"github.com/egoist/mygo/ui"
)

// newTestApp 返回一个空的 App（只有今天这一天）与它的数据目录。
func newTestApp(t *testing.T) (*App, string) {
	t.Helper()
	dir := t.TempDir()
	s, err := store.New(dir)
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	return NewApp(s), dir
}

// TestAddDiaryEntryThroughUI 用真实的控件走一遍「点开弹窗 → 填表 → 保存」。
func TestAddDiaryEntryThroughUI(t *testing.T) {
	app, dir := newTestApp(t)
	app.router.Push("/diary")

	tt := ui.NewTester(app.View, 1200, 820)
	if err := tt.Click("写日记"); err != nil {
		t.Fatalf("点「写日记」: %v", err)
	}
	if !app.ed.Open {
		t.Fatalf("弹窗没有打开")
	}

	// 科目拿到了 AutoFocus，直接打字即可。
	tt.Type("计算机组成原理")
	tt.Click("章节")
	tt.Type("第 4 章 存储系统 >> 第 8 节 虚拟存储器")

	tt.Click("要点")
	tt.Type("4.8.2 页式虚拟存储器 -（2）地址映射和页表")
	tt.Key(0, ui.KeyEnter) // 多行输入框里回车就是换行
	tt.Type("- 4.8.2 页式虚拟存储器 -（3）访问流程")

	if err := tt.Click("保存"); err != nil {
		t.Fatalf("点「保存」: %v", err)
	}
	if app.ed.Open {
		t.Errorf("保存后弹窗应该关闭, err = %q", app.ed.Err)
	}

	day := app.day(app.today)
	if len(day.Entries) != 1 {
		t.Fatalf("条目数 = %d, 期望 1", len(day.Entries))
	}
	e := day.Entries[0]
	if e.Subject != "计算机组成原理" {
		t.Errorf("科目 = %q", e.Subject)
	}
	if len(e.Points) != 2 {
		t.Fatalf("要点数 = %d, 期望 2: %+v", len(e.Points), e.Points)
	}
	if e.Points[1].Text != "4.8.2 页式虚拟存储器 -（3）访问流程" {
		t.Errorf("第二行开头的 - 应该被去掉, 得到 %q", e.Points[1].Text)
	}

	// 真的落盘了，而且文件名就是日期。
	raw, err := os.ReadFile(filepath.Join(dir, app.today+".json"))
	if err != nil {
		t.Fatalf("读落盘文件: %v", err)
	}
	if !strings.Contains(string(raw), "计算机组成原理") {
		t.Errorf("文件里没有刚写的内容: %s", raw)
	}
	if !strings.Contains(string(raw), "地址映射和页表") {
		t.Errorf("文件里没有要点: %s", raw)
	}
}

// TestEditorValidation 检查「什么都不填就保存」会被拦下来。
func TestEditorValidation(t *testing.T) {
	app, _ := newTestApp(t)
	app.router.Push("/diary")
	tt := ui.NewTester(app.View, 1200, 820)

	if err := tt.Click("写日记"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("保存"); err != nil {
		t.Fatal(err)
	}
	if !app.ed.Open {
		t.Errorf("校验没通过时不应该关掉弹窗")
	}
	if app.ed.Err == "" {
		t.Errorf("应该给出错误提示")
	}
	if !tt.HasText("写一下科目吧") {
		t.Errorf("错误提示没有显示在界面上: %q", tt.Texts())
	}
	if len(app.day(app.today).Entries) != 0 {
		t.Errorf("不该写入任何记录")
	}
}

// TestCheckPointWritesThrough 打勾一个要点，值要写进 model 并落盘。
func TestCheckPointWritesThrough(t *testing.T) {
	app, dir := newTestApp(t)
	app.router.Push("/diary")
	// 先造一条今天的记录
	app.openEditor(app.today, "")
	app.ed.Subject = "数据结构"
	app.ed.Points = "图的基本概念\n邻接矩阵"
	app.saveEditor()

	tt := ui.NewTester(app.View, 1200, 820)
	if err := tt.Click("图的基本概念"); err != nil {
		t.Fatalf("点要点: %v", err)
	}
	if !app.day(app.today).Entries[0].Points[0].Done {
		t.Fatalf("要点没有被打勾: %+v", app.day(app.today).Entries[0].Points)
	}

	raw, err := os.ReadFile(filepath.Join(dir, app.today+".json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"done": true`) {
		t.Errorf("打勾没有落盘: %s", raw)
	}
}

// TestTodoFlow 覆盖添加、勾选、删除、撤销。
func TestTodoFlow(t *testing.T) {
	app, dir := newTestApp(t)
	app.router.Push("/todo")
	tt := ui.NewTester(app.View, 1200, 820)

	// 添加：输入后回车
	if err := tt.Click("新待办"); err != nil {
		t.Fatalf("点输入框: %v", err)
	}
	tt.Type("复习 B 树")
	tt.Key(0, ui.KeyEnter)
	if got := len(app.day(app.today).Todos); got != 1 {
		t.Fatalf("回车后待办数 = %d, 期望 1", got)
	}
	if app.todoDraft != "" {
		t.Errorf("添加后输入框应清空, 得到 %q", app.todoDraft)
	}

	// 勾选
	if err := tt.Click("复习 B 树"); err != nil {
		t.Fatalf("点待办: %v", err)
	}
	if !app.day(app.today).Todos[0].Done {
		t.Fatalf("待办没有被打勾")
	}

	// 删除（按钮只在 hover 时出现，这里直接调方法，避免依赖指针位置）
	id := app.day(app.today).Todos[0].ID
	app.deleteTodo(id)
	if len(app.day(app.today).Todos) != 0 {
		t.Fatalf("删除后还有 %d 条", len(app.day(app.today).Todos))
	}
	if len(app.toasts) == 0 || app.toasts[len(app.toasts)-1].Action != "撤销" {
		t.Fatalf("删除后应该有一条可撤销的提示: %+v", app.toasts)
	}

	// 撤销
	app.toasts[len(app.toasts)-1].OnAct()
	if got := len(app.day(app.today).Todos); got != 1 {
		t.Fatalf("撤销后待办数 = %d, 期望 1", got)
	}
	if !app.day(app.today).Todos[0].Done {
		t.Errorf("撤销应该把「已完成」状态一起带回来")
	}

	raw, err := os.ReadFile(filepath.Join(dir, app.today+".json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "复习 B 树") {
		t.Errorf("撤销后文件里应该有这条待办: %s", raw)
	}
}

// TestTodoFilter 检查「进行中 / 已完成」两个筛选。
func TestTodoFilter(t *testing.T) {
	app, _ := newTestApp(t)
	app.days[0].Todos = []model.Todo{
		{ID: "a", Text: "没做的"},
		{ID: "b", Text: "做完的", Done: true},
	}

	app.todoFilter = 0
	if got := len(app.todoRowsForToday()); got != 2 {
		t.Errorf("全部 = %d, 期望 2", got)
	}
	app.todoFilter = 1
	if rows := app.todoRowsForToday(); len(rows) != 1 || app.days[0].Todos[rows[0]].ID != "a" {
		t.Errorf("进行中 = %v", rows)
	}
	app.todoFilter = 2
	if rows := app.todoRowsForToday(); len(rows) != 1 || app.days[0].Todos[rows[0]].ID != "b" {
		t.Errorf("已完成 = %v", rows)
	}
}

// TestDiarySearch 检查搜索会命中科目、要点、备注。
func TestDiarySearch(t *testing.T) {
	app := sampleApp(t)
	app.router.Push("/diary")
	tt := ui.NewTester(app.View, 1200, 820)

	if err := tt.Click("搜索日记"); err != nil {
		t.Fatalf("点搜索框: %v", err)
	}
	tt.Type("邻接矩阵")
	if rows := len(app.filterDays()); rows != 1 {
		t.Fatalf("搜「邻接矩阵」命中 %d 天, 期望 1", rows)
	}
	if !tt.HasText("11 月 13 日") {
		t.Errorf("搜索结果里应该有今天: %q", tt.Texts())
	}

	// 清空后恢复显示所有日子
	app.query = ""
	if rows := len(app.filterDays()); rows != len(app.days) {
		t.Errorf("清空搜索后 %d 天, 期望 %d 天", rows, len(app.days))
	}

	// 搜不到时给空状态
	app.query = "根本不存在的关键词"
	tt.Frame()
	if !tt.HasText("没有找到相关记录") {
		t.Errorf("没有空状态提示: %q", tt.Texts())
	}
}

// TestThemePrefsPersist 检查外观设置写进了 settings.json。
func TestThemePrefsPersist(t *testing.T) {
	app, dir := newTestApp(t)
	app.themePick = 2
	app.accentPick = 2
	app.saveSettings()

	st, err := app.store.LoadSettings()
	if err != nil {
		t.Fatalf("LoadSettings: %v", err)
	}
	if st.Theme != model.ThemeDark {
		t.Errorf("Theme = %q, 期望 dark", st.Theme)
	}
	if st.Accent != accentChoices[2].Hex {
		t.Errorf("Accent = %q, 期望 %q", st.Accent, accentChoices[2].Hex)
	}
	if _, err := os.Stat(filepath.Join(dir, "settings.json")); err != nil {
		t.Errorf("settings.json 不存在: %v", err)
	}
}

// TestSwitchAppearanceThroughUI 用界面里的分段控件真的切一次深色。
func TestSwitchAppearanceThroughUI(t *testing.T) {
	app, _ := newTestApp(t)
	tt := ui.NewTester(app.View, 1200, 820)

	if err := tt.Click("外观 · 跟随系统"); err != nil {
		t.Fatalf("点开外观: %v", err)
	}
	if !app.appearance {
		t.Fatalf("外观面板没有打开")
	}
	if err := tt.Click("深色"); err != nil {
		t.Fatalf("点深色: %v", err)
	}
	if app.themePick != 2 {
		t.Fatalf("themePick = %d, 期望 2", app.themePick)
	}
	st, err := app.store.LoadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if st.Theme != model.ThemeDark {
		t.Errorf("设置没有落盘: %+v", st)
	}
}

// TestDeleteDayWithUndo 清空一天，再从提示里撤销回来。
func TestDeleteDayWithUndo(t *testing.T) {
	app := sampleApp(t)
	date := model.MustDate(model.Today()).AddDate(0, 0, -1).Format(model.DateLayout)
	before := app.day(date)
	if !before.HasContent() {
		t.Fatalf("样本数据里昨天应该有内容")
	}

	app.deleteDay(date)
	if app.day(date).HasContent() {
		t.Errorf("清空后不应有内容")
	}
	// 撤销
	if len(app.toasts) == 0 {
		t.Fatalf("没有撤销提示")
	}
	app.toasts[len(app.toasts)-1].OnAct()
	after := app.day(date)
	if len(after.Entries) != len(before.Entries) || len(after.Todos) != len(before.Todos) {
		t.Errorf("撤销后 = %d 条记录 / %d 条待办, 期望 %d / %d",
			len(after.Entries), len(after.Todos), len(before.Entries), len(before.Todos))
	}
}

// TestStreak 检查连续记录的计算：今天没写不算断。
func TestStreak(t *testing.T) {
	app, _ := newTestApp(t)
	today := model.MustDate(model.Today())

	// 今天空着，昨天、前天有内容 → 连续 2 天
	for i := 1; i <= 2; i++ {
		d := today.AddDate(0, 0, -i).Format(model.DateLayout)
		app.upsert(model.Day{Date: d, Todos: []model.Todo{{ID: d, Text: "x"}}})
	}
	if got := app.streak(); got != 2 {
		t.Errorf("streak = %d, 期望 2", got)
	}

	// 今天也写 → 3
	app.upsert(model.Day{Date: app.today, Todos: []model.Todo{{ID: "t", Text: "y"}}})
	if got := app.streak(); got != 3 {
		t.Errorf("streak = %d, 期望 3", got)
	}

	// 中间断一天 → 只算到今天
	app.days = app.days[:0]
	app.ensureToday()
	app.upsert(model.Day{Date: app.today, Todos: []model.Todo{{ID: "t", Text: "y"}}})
	if got := app.streak(); got != 1 {
		t.Errorf("断了之后 streak = %d, 期望 1", got)
	}
}

// TestParsePointLines 检查从旧 Markdown 粘贴的列表能被正确拆开。
func TestParsePointLines(t *testing.T) {
	got := parsePointLines("- 第一条\n\n* 第二条\n• 第三条\n\n  第四条  \n")
	want := []string{"第一条", "第二条", "第三条", "第四条"}
	if len(got) != len(want) {
		t.Fatalf("拆出 %d 条, 期望 %d 条: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i].Text != want[i] {
			t.Errorf("第 %d 条 = %q, 期望 %q", i, got[i].Text, want[i])
		}
	}
	if got[0].ID == "" || got[1].ID == "" || got[0].ID == got[1].ID {
		t.Errorf("要点应该有各自的 ID: %+v", got)
	}
}

// TestTopicParts 检查章节路径的拆分，顺便确认不拆坏 TCP/IP 这种名字。
func TestTopicParts(t *testing.T) {
	cases := map[string][]string{
		"第 4 章 存储系统 >> 第 8 节 虚拟存储器": {"第 4 章 存储系统", "第 8 节 虚拟存储器"},
		"图 › 最短路":     {"图", "最短路"},
		"TCP/IP 协议栈":  {"TCP/IP 协议栈"},
		"":            nil,
		">> 只有分隔符 >>": {"只有分隔符"},
	}
	for in, want := range cases {
		got := model.TopicParts(in)
		if len(got) != len(want) {
			t.Errorf("TopicParts(%q) = %v, 期望 %v", in, got, want)
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("TopicParts(%q)[%d] = %q, 期望 %q", in, i, got[i], want[i])
			}
		}
	}
}

// TestShortcutsAreInertBehindModal 确认弹窗打开时，窗口级快捷键不会误触发。
//
// 这一条很容易被写错：如果 ⌘N 在弹窗里还能用，用户正在写的草稿会被清空。
func TestShortcutsAreInertBehindModal(t *testing.T) {
	app, _ := newTestApp(t)
	app.router.Push("/diary")
	tt := ui.NewTester(app.View, 1200, 820)

	if err := tt.Click("写日记"); err != nil {
		t.Fatal(err)
	}
	tt.Type("数据结构")
	if app.ed.Subject != "数据结构" {
		t.Fatalf("科目 = %q", app.ed.Subject)
	}

	tt.Key(ui.Cmd, ui.KeyN) // 不该把草稿清空
	if app.ed.Subject != "数据结构" {
		t.Errorf("⌘N 在弹窗里清空了草稿: %q", app.ed.Subject)
	}
	tt.Key(ui.Cmd, ui.Key3) // 不该切页面
	if got := app.router.Path(); got != "/diary" {
		t.Errorf("⌘3 在弹窗里切走了页面: %q", got)
	}
}

func TestEmptyDayNeverShowsAnOldCard(t *testing.T) {
	app := sampleApp(t)
	date := model.MustDate(model.Today()).AddDate(0, 0, -3).Format(model.DateLayout)

	app.deleteDay(date)
	app.toasts = nil // 不要撤销
	for _, d := range app.days {
		if d.Date == date && !d.HasContent() {
			t.Errorf("空的 %s 还留在列表里", date)
		}
	}
	if _, err := os.Stat(filepath.Join(app.store.Dir(), date+".json")); !os.IsNotExist(err) {
		t.Errorf("空日子的文件应该被删掉, stat err = %v", err)
	}
}
