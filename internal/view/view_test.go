package view

import (
	"path/filepath"
	"strings"
	"testing"

	"flowlog-desktop/internal/model"
	"flowlog-desktop/internal/store"

	"github.com/egoist/mygo/ui"
)

// newTestApp 返回一个空的 App（只有今天这一天）与它的数据库。
func newTestApp(t *testing.T) *App {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), store.FileName))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	if err := s.SaveSettings(model.Settings{Theme: model.ThemeLight, Onboarded: true}); err != nil {
		t.Fatal(err)
	}
	app, err := NewApp(s)
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}
	return app
}

// TestAddDiaryEntryThroughUI 用真实的控件走一遍「点开弹窗 → 填表 → 保存」。
func TestAddDiaryEntryThroughUI(t *testing.T) {
	app := newTestApp(t)
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
	tt.Click("章")
	tt.Type("第 4 章 存储系统")
	tt.Click("节")
	tt.Type("第 8 节 虚拟存储器")

	tt.Click("要点")
	tt.Type("页式虚拟存储器的地址映射和页表")
	tt.Key(0, ui.KeyEnter) // 多行输入框里回车就是换行
	tt.Type("- 缺页中断的处理流程")

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
	if e.Chapter != "第 4 章 存储系统" || e.Section != "第 8 节 虚拟存储器" {
		t.Errorf("章/节 = %q / %q", e.Chapter, e.Section)
	}
	if len(e.Points) != 2 {
		t.Fatalf("要点数 = %d, 期望 2: %+v", len(e.Points), e.Points)
	}
	if e.Points[1].Text != "缺页中断的处理流程" {
		t.Errorf("第二行开头的 - 应该被去掉, 得到 %q", e.Points[1].Text)
	}

	// 真的落库了：重新打开数据库还能读到。
	reopened, err := store.Open(app.store.Path())
	if err != nil {
		t.Fatalf("重新打开数据库: %v", err)
	}
	defer reopened.Close()
	days, err := reopened.Days()
	if err != nil {
		t.Fatal(err)
	}
	if len(days) != 1 || days[0].Entries[0].Chapter != "第 4 章 存储系统" {
		t.Fatalf("落库内容 = %+v", days)
	}
}

// TestEditorValidation 检查「什么都不填就保存」会被拦下来。
func TestEditorValidation(t *testing.T) {
	app := newTestApp(t)
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
	if !tt.HasText("请填写科目") {
		t.Errorf("错误提示没有显示在界面上: %q", tt.Texts())
	}
	if len(app.day(app.today).Entries) != 0 {
		t.Errorf("不该写入任何记录")
	}
}

// TestCheckPointWritesThrough 打勾一个要点，值要写进 model 并落库。
func TestCheckPointWritesThrough(t *testing.T) {
	app := newTestApp(t)
	app.router.Push("/diary")
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

	// 重新读库确认写进去了。
	reopened, err := store.Open(app.store.Path())
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	days, _ := reopened.Days()
	if !days[0].Entries[0].Points[0].Done {
		t.Errorf("打勾没有落库: %+v", days[0].Entries[0].Points)
	}
}

// TestTodoFlow 覆盖添加、勾选、删除、撤销。
func TestTodoFlow(t *testing.T) {
	app := newTestApp(t)
	app.router.Push("/todo")
	tt := ui.NewTester(app.View, 1200, 820)

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

	if err := tt.Click("复习 B 树"); err != nil {
		t.Fatalf("点待办: %v", err)
	}
	if !app.day(app.today).Todos[0].Done {
		t.Fatalf("待办没有被打勾")
	}

	id := app.day(app.today).Todos[0].ID
	app.deleteTodo(id)
	if len(app.day(app.today).Todos) != 0 {
		t.Fatalf("删除后还有 %d 条", len(app.day(app.today).Todos))
	}
	if len(app.toasts) == 0 || app.toasts[len(app.toasts)-1].Action != "撤销" {
		t.Fatalf("删除后应该有一条可撤销的提示: %+v", app.toasts)
	}

	app.toasts[len(app.toasts)-1].OnAct()
	if got := len(app.day(app.today).Todos); got != 1 {
		t.Fatalf("撤销后待办数 = %d, 期望 1", got)
	}
	if !app.day(app.today).Todos[0].Done {
		t.Errorf("撤销应该把「已完成」状态一起带回来")
	}
}

// TestTodoFilter 检查「进行中 / 已完成」两个筛选。
func TestTodoFilter(t *testing.T) {
	app := newTestApp(t)
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

// TestDiarySearch 检查搜索会命中科目、章节与要点。
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

	app.query = ""
	if rows := len(app.filterDays()); rows != len(app.days) {
		t.Errorf("清空搜索后 %d 天, 期望 %d 天", rows, len(app.days))
	}

	app.query = "不存在的关键词"
	tt.Frame()
	if !tt.HasText("没有匹配的记录") {
		t.Errorf("没有空状态提示: %q", tt.Texts())
	}
}

// TestThemePersists 检查外观写进了数据库。
func TestThemePersists(t *testing.T) {
	app := newTestApp(t)
	app.settings.Theme = model.ThemeDark
	app.saveSettings()

	reopened, err := store.Open(app.store.Path())
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	st, err := reopened.Settings()
	if err != nil {
		t.Fatal(err)
	}
	if st.Theme != model.ThemeDark {
		t.Errorf("Theme = %q, 期望 dark", st.Theme)
	}
}

// TestSwitchThemeThroughUI 用设置页的分段控件真的切一次深色。
func TestSwitchThemeThroughUI(t *testing.T) {
	app := newTestApp(t)
	app.router.Push("/settings")
	tt := ui.NewTester(app.View, 1200, 820)

	if err := tt.Click("深色"); err != nil {
		t.Fatalf("点深色: %v", err)
	}
	if app.settings.Theme != model.ThemeDark {
		t.Fatalf("Theme = %q, 期望 dark", app.settings.Theme)
	}
	st, err := app.store.Settings()
	if err != nil {
		t.Fatal(err)
	}
	if st.Theme != model.ThemeDark {
		t.Errorf("设置没有落库: %+v", st)
	}
}

// TestOnboarding 检查首次启动的称呼引导。
func TestOnboarding(t *testing.T) {
	s, err := store.Open(filepath.Join(t.TempDir(), store.FileName))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	app, err := NewApp(s)
	if err != nil {
		t.Fatal(err)
	}
	app.Start()
	if !app.onboarding.Open {
		t.Fatalf("首次启动应该打开称呼引导")
	}

	tt := ui.NewTester(app.View, 1200, 820)
	if err := tt.Click("开始使用"); err != nil {
		t.Fatal(err)
	}
	if app.onboarding.Err == "" {
		t.Errorf("空称呼应该提示错误")
	}

	tt.Type("小明")
	if err := tt.Click("开始使用"); err != nil {
		t.Fatal(err)
	}
	if app.onboarding.Open {
		t.Errorf("填了称呼之后引导应该关闭")
	}
	if app.settings.UserName != "小明" || !app.settings.Onboarded {
		t.Errorf("设置 = %+v", app.settings)
	}
	if !tt.HasText("晚上好，小明") && !strings.Contains(strings.Join(tt.Texts(), " "), "小明") {
		t.Errorf("首页问候里应该有称呼: %q", tt.Texts())
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
	app := newTestApp(t)
	today := model.MustDate(model.Today())

	for i := 1; i <= 2; i++ {
		d := today.AddDate(0, 0, -i).Format(model.DateLayout)
		app.upsert(model.Day{Date: d, Todos: []model.Todo{{ID: d, Text: "x"}}})
	}
	if got := app.streak(); got != 2 {
		t.Errorf("streak = %d, 期望 2", got)
	}

	app.upsert(model.Day{Date: app.today, Todos: []model.Todo{{ID: "t", Text: "y"}}})
	if got := app.streak(); got != 3 {
		t.Errorf("streak = %d, 期望 3", got)
	}

	app.days = app.days[:0]
	app.ensureToday()
	app.upsert(model.Day{Date: app.today, Todos: []model.Todo{{ID: "t", Text: "y"}}})
	if got := app.streak(); got != 1 {
		t.Errorf("断了之后 streak = %d, 期望 1", got)
	}
}

// TestShortcutsAreInertBehindModal 确认弹窗打开时，窗口级快捷键不会误触发。
func TestShortcutsAreInertBehindModal(t *testing.T) {
	app := newTestApp(t)
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
	if got[0].ID == "" || got[0].ID == got[1].ID {
		t.Errorf("要点应该有各自的 ID: %+v", got)
	}
}

// TestTopicParts 检查章 / 节的展示，顺便确认兼容老的 ">>" 写法。
func TestTopicParts(t *testing.T) {
	cases := []struct {
		chapter, section string
		want             []string
	}{
		{"第 4 章 存储系统", "第 8 节 虚拟存储器", []string{"第 4 章 存储系统", "第 8 节 虚拟存储器"}},
		{"第 4 章 存储系统", "", []string{"第 4 章 存储系统"}},
		{"", "", nil},
		// 老数据把两段塞在一个字段里
		{"第 4 章 存储系统 >> 第 8 节 虚拟存储器", "", []string{"第 4 章 存储系统", "第 8 节 虚拟存储器"}},
		// 不拆坏 TCP/IP
		{"TCP/IP 协议栈", "", []string{"TCP/IP 协议栈"}},
	}
	for _, c := range cases {
		got := model.Entry{Chapter: c.chapter, Section: c.section}.TopicParts()
		if len(got) != len(c.want) {
			t.Errorf("TopicParts(%q, %q) = %v, 期望 %v", c.chapter, c.section, got, c.want)
			continue
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Errorf("TopicParts(%q, %q)[%d] = %q, 期望 %q", c.chapter, c.section, i, got[i], c.want[i])
			}
		}
	}
}

// TestEmptyDayNeverShowsAnOldCard 清空一天之后，列表里不该留一张空卡片。
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
	days, err := app.store.Days()
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range days {
		if d.Date == date {
			t.Errorf("数据库里还留着空的一天 %s", date)
		}
	}
}
