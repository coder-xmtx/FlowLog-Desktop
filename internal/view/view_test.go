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
	if err := s.SaveSettings(model.Settings{Theme: model.ThemeDark, Onboarded: true}); err != nil {
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

// TestEditorValidationShownOutsideScroll 校验错误要出现在滚动区外，
// 用户不用滚动就能看到。
func TestEditorValidationShownOutsideScroll(t *testing.T) {
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

// TestTodoFlow 覆盖添加（带截止日期）、勾选、删除、撤销。
func TestTodoFlow(t *testing.T) {
	app := newTestApp(t)
	app.router.Push("/todo")
	tt := ui.NewTester(app.View, 1200, 820)

	if err := tt.Click("新待办"); err != nil {
		t.Fatalf("点输入框: %v", err)
	}
	tt.Type("复习 B 树")

	// 打开截止日期弹层，选「明天」。
	if err := tt.Click("选择截止日期"); err != nil {
		t.Fatalf("点截止日期: %v", err)
	}
	if !app.duePicker {
		t.Fatalf("截止日期弹层没有打开")
	}
	if err := tt.Click("明天"); err != nil {
		t.Fatalf("点「明天」: %v", err)
	}
	if got := app.todoDue.Format(model.DateLayout); got != model.DayAfter(app.today, 1) {
		t.Fatalf("截止日期 = %q, 期望明天", got)
	}

	tt.Key(0, ui.KeyEnter)
	if len(app.todos) != 1 {
		t.Fatalf("回车后待办数 = %d, 期望 1", len(app.todos))
	}
	todo := app.todos[0]
	if todo.Text != "复习 B 树" {
		t.Errorf("内容 = %q", todo.Text)
	}
	if todo.Due != model.DayAfter(app.today, 1) {
		t.Errorf("截止日期没保存: %+v", todo)
	}
	if todo.Created != app.today {
		t.Errorf("创建日期 = %q", todo.Created)
	}
	if app.todoDraft != "" {
		t.Errorf("添加后输入框应清空, 得到 %q", app.todoDraft)
	}
	if !app.todoDue.IsZero() {
		t.Errorf("添加后截止日期应重置")
	}

	// 勾选
	if err := tt.Click("复习 B 树"); err != nil {
		t.Fatalf("点待办: %v", err)
	}
	if !app.todos[0].Done {
		t.Fatalf("待办没有被打勾")
	}

	id := app.todos[0].ID
	app.deleteTodo(id)
	if len(app.todos) != 0 {
		t.Fatalf("删除后还有 %d 条", len(app.todos))
	}
	if len(app.toasts) == 0 || app.toasts[len(app.toasts)-1].Action != "撤销" {
		t.Fatalf("删除后应该有一条可撤销的提示: %+v", app.toasts)
	}

	app.toasts[len(app.toasts)-1].OnAct()
	if len(app.todos) != 1 {
		t.Fatalf("撤销后待办数 = %d, 期望 1", len(app.todos))
	}
	if !app.todos[0].Done {
		t.Errorf("撤销应该把「已完成」状态一起带回来")
	}
}

// TestTodoFilter 检查「进行中 / 已完成」两个筛选。
func TestTodoFilter(t *testing.T) {
	app := newTestApp(t)
	app.todos = []model.Todo{
		{ID: "a", Text: "没做的"},
		{ID: "b", Text: "做完的", Done: true},
	}

	app.todoFilter = 0
	if got := len(app.todoRows()); got != 2 {
		t.Errorf("全部 = %d, 期望 2", got)
	}
	app.todoFilter = 1
	if rows := app.todoRows(); len(rows) != 1 || app.todos[rows[0]].ID != "a" {
		t.Errorf("进行中 = %v", rows)
	}
	app.todoFilter = 2
	if rows := app.todoRows(); len(rows) != 1 || app.todos[rows[0]].ID != "b" {
		t.Errorf("已完成 = %v", rows)
	}
}

// TestTodoSeparatedFromDiary 确认待办与日记解耦：
// 只加待办不会让日记的天数、今天的记录数变化，删日记也不影响待办。
func TestTodoSeparatedFromDiary(t *testing.T) {
	app := newTestApp(t)
	before := app.daysWritten()
	todayBefore := app.todayEntryCount()

	app.openEditor(app.today, "")
	app.ed.Subject = "数据结构"
	app.ed.Points = "图的遍历"
	app.saveEditor()
	if got := app.todayEntryCount(); got != todayBefore+1 {
		t.Fatalf("写日记后今日记录 = %d", got)
	}

	app.todoDraft = "只加一条待办"
	app.addTodo()
	if len(app.todos) != 1 {
		t.Fatalf("待办数 = %d", len(app.todos))
	}
	if got := app.daysWritten(); got != before+1 {
		t.Errorf("加待办不该改变日记天数: %d → %d", before, got)
	}
	if got := app.todayEntryCount(); got != todayBefore+1 {
		t.Errorf("加待办不该改变今日记录数: %d", got)
	}

	// 删掉这一天的日记，待办必须还在。
	app.deleteDay(app.today)
	if len(app.day(app.today).Entries) != 0 {
		t.Errorf("日记没被清空")
	}
	if len(app.todos) != 1 {
		t.Errorf("删日记把待办也删了: %+v", app.todos)
	}
}

// TestTodoDueOrder 确认列表按截止日期排序，没有截止的排最后。
func TestTodoDueOrder(t *testing.T) {
	app := newTestApp(t)
	today := app.today
	app.todos = nil
	for _, todo := range []model.Todo{
		{ID: "none", Text: "无截止", Created: today},
		{ID: "later", Text: "以后", Created: today, Due: model.DayAfter(today, 5)},
		{ID: "tomorrow", Text: "明天", Created: today, Due: model.DayAfter(today, 1)},
		{ID: "done", Text: "已完成", Created: today, Done: true, Due: model.DayAfter(today, 0)},
	} {
		app.saveTodo(todo)
	}
	want := []string{"tomorrow", "later", "none", "done"}
	for i, id := range want {
		if app.todos[i].ID != id {
			t.Fatalf("第 %d 条 = %q, 期望 %q", i, app.todos[i].ID, id)
		}
	}
}

// TestHomeDueCard 确认首页只展示过期、今天、明天到期的待办。
func TestHomeDueCard(t *testing.T) {
	app := sampleApp(t)
	app.router.Push("/")
	tt := ui.NewTester(app.View, 1200, 820)

	texts := strings.Join(tt.Texts(), "\n")
	for _, want := range []string{
		"待办提醒",
		"复习 B 树的分裂与合并", // 明天到期
		"交操作系统作业",      // 已过期
		"明天到期",
		"已过期",
	} {
		if !strings.Contains(texts, want) {
			t.Errorf("首页缺少 %q:\n%s", want, texts)
		}
	}
	// 没有截止日期的不该出现在提醒里。
	if strings.Contains(texts, "整理虚拟存储器的笔记") {
		t.Errorf("没有截止日期的待办不该出现在到期提醒里")
	}
	// 已完成的也不该出现。
	if strings.Contains(texts, "做一章 408 真题") {
		t.Errorf("已完成的待办不该出现在到期提醒里")
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
	if !tt.HasText("11 月 13 日") && !tt.HasText(model.DisplayDate(app.today)) {
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
	app.settings.Theme = model.ThemeSystem
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
	if st.Theme != model.ThemeSystem {
		t.Errorf("Theme = %q, 期望 system", st.Theme)
	}
}

// TestSwitchThemeThroughUI 用设置页的分段控件真的切一次浅色。
func TestSwitchThemeThroughUI(t *testing.T) {
	app := newTestApp(t)
	app.router.Push("/settings")
	tt := ui.NewTester(app.View, 1200, 820)

	if err := tt.Click("浅色"); err != nil {
		t.Fatalf("点浅色: %v", err)
	}
	if app.settings.Theme != model.ThemeLight {
		t.Fatalf("Theme = %q, 期望 light", app.settings.Theme)
	}
	st, err := app.store.Settings()
	if err != nil {
		t.Fatal(err)
	}
	if st.Theme != model.ThemeLight {
		t.Errorf("设置没有落库: %+v", st)
	}

	// 跟随系统
	if err := tt.Click("跟随系统"); err != nil {
		t.Fatalf("点跟随系统: %v", err)
	}
	if app.settings.Theme != model.ThemeSystem {
		t.Errorf("Theme = %q, 期望 system", app.settings.Theme)
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
}

// TestDeleteDayWithUndo 清空一天，再从提示里撤销回来。
func TestDeleteDayWithUndo(t *testing.T) {
	app := sampleApp(t)
	date := model.MustDate(model.Today()).AddDate(0, 0, -1).Format(model.DateLayout)
	before := app.day(date)
	if !before.HasEntries() {
		t.Fatalf("样本数据里昨天应该有内容")
	}

	app.deleteDay(date)
	if app.day(date).HasEntries() {
		t.Errorf("清空后不应有内容")
	}
	if len(app.toasts) == 0 {
		t.Fatalf("没有撤销提示")
	}
	app.toasts[len(app.toasts)-1].OnAct()
	after := app.day(date)
	if len(after.Entries) != len(before.Entries) {
		t.Errorf("撤销后 = %d 条记录, 期望 %d", len(after.Entries), len(before.Entries))
	}
}

// TestStreak 检查连续记录的计算：今天没写不算断。
func TestStreak(t *testing.T) {
	app := newTestApp(t)
	today := model.MustDate(model.Today())

	for i := 1; i <= 2; i++ {
		d := today.AddDate(0, 0, -i).Format(model.DateLayout)
		app.upsert(model.Day{Date: d, Entries: []model.Entry{{ID: d, Subject: "x"}}})
	}
	if got := app.streak(); got != 2 {
		t.Errorf("streak = %d, 期望 2", got)
	}

	app.upsert(model.Day{Date: app.today, Entries: []model.Entry{{ID: "t", Subject: "y"}}})
	if got := app.streak(); got != 3 {
		t.Errorf("streak = %d, 期望 3", got)
	}

	app.days = app.days[:0]
	app.ensureToday()
	app.upsert(model.Day{Date: app.today, Entries: []model.Entry{{ID: "t", Subject: "y"}}})
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
		{"第 4 章 存储系统 >> 第 8 节 虚拟存储器", "", []string{"第 4 章 存储系统", "第 8 节 虚拟存储器"}},
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

// TestEmptyDayNeverShowsAnOldCard 清空一天之后，列表与数据库里都不该留空壳。
func TestEmptyDayNeverShowsAnOldCard(t *testing.T) {
	app := sampleApp(t)
	date := model.MustDate(model.Today()).AddDate(0, 0, -3).Format(model.DateLayout)

	app.deleteDay(date)
	app.toasts = nil // 不要撤销
	for _, d := range app.days {
		if d.Date == date && !d.HasEntries() {
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

// TestDueSoonModel 直接测「近两天到期」的筛选与排序。
func TestDueSoonModel(t *testing.T) {
	today := "2026-11-14"
	todos := []model.Todo{
		{ID: "future", Due: "2026-11-20"},
		{ID: "tomorrow", Due: "2026-11-15", Created: "2026-11-10"},
		{ID: "today", Due: "2026-11-14", Created: "2026-11-10"},
		{ID: "overdue", Due: "2026-11-12", Created: "2026-11-09"},
		{ID: "none", Created: "2026-11-13"},
		{ID: "done", Due: "2026-11-14", Done: true},
	}
	got := model.DueSoon(todos, today)
	want := []string{"overdue", "today", "tomorrow"}
	if len(got) != len(want) {
		t.Fatalf("DueSoon = %v", ids(got))
	}
	for i := range want {
		if got[i].ID != want[i] {
			t.Errorf("第 %d 条 = %q, 期望 %q", i, got[i].ID, want[i])
		}
	}
	if got[0].DueStateOf(today) != model.DueOverdue {
		t.Errorf("状态 = %v", got[0].DueStateOf(today))
	}
}

func ids(todos []model.Todo) []string {
	out := make([]string, len(todos))
	for i, t := range todos {
		out[i] = t.ID
	}
	return out
}

// TestReleaseNotesExtract 检查从 CHANGELOG 里取版本说明。
func TestReleaseNotesExtract(t *testing.T) {
	md := "# 更新日志\n\n## 0.5.0\n\n- 新功能 A\n- 修复 B\n\n## 0.4.0\n\n- 旧版本的说明\n"
	got := extractReleaseNotes(md, "0.5.0")
	if !strings.Contains(got, "新功能 A") || strings.Contains(got, "旧版本") {
		t.Errorf("extractReleaseNotes = %q", got)
	}
	if extractReleaseNotes(md, "9.9.9") != "" {
		t.Errorf("不存在的版本应该返回空")
	}
	if !strings.Contains(notesText(got, 5), "· 新功能 A") {
		t.Errorf("notesText = %q", notesText(got, 5))
	}
}
