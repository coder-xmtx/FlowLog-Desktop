package view

import (
	"strings"
	"testing"

	"flowlog-desktop/internal/app"
	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// newTester 打开主界面的测试器。
func newTester(t *testing.T, a *app.App, w, h int) *ui.Tester {
	t.Helper()
	tt := ui.NewTester(func(c *ui.Context) { Root(c, a) }, w, h)
	tt.Frame()
	return tt
}

// TestAddStudyEntryThroughUI 用真实控件走一遍「点写学记 → 填表 → 保存」。
func TestAddStudyEntryThroughUI(t *testing.T) {
	a := sampleApp(t)
	a.Router.Push("/study")
	tt := newTester(t, a, 1200, 820)

	if err := tt.Click("写学记"); err != nil {
		t.Fatalf("点「写学记」: %v", err)
	}
	before := len(a.Day(a.Today).Entries)
	if !a.Editor.Open {
		t.Fatalf("弹窗没有打开")
	}

	// 科目拿到了 AutoFocus，直接打字即可。
	tt.Type("数据结构")
	tt.Click("章")
	tt.Type("第 6 章 图")
	tt.Click("要点")
	tt.Type("图的基本概念")
	tt.Key(0, ui.KeyEnter)
	tt.Type("邻接矩阵与邻接表")

	if err := tt.Click("保存"); err != nil {
		t.Fatalf("点「保存」: %v", err)
	}
	if a.Editor.Open {
		t.Fatalf("保存后弹窗应该关闭，Err=%q", a.Editor.Err)
	}

	day := a.Day(a.Today)
	if len(day.Entries) != before+1 {
		t.Fatalf("条目数 = %d, 期望 %d", len(day.Entries), before+1)
	}
	e := day.Entries[len(day.Entries)-1]
	if e.Subject != "数据结构" || e.Chapter != "第 6 章 图" {
		t.Errorf("科目/章 = %q / %q", e.Subject, e.Chapter)
	}
	if len(e.Points) != 2 {
		t.Fatalf("要点数 = %d: %+v", len(e.Points), e.Points)
	}
	if e.Points[1].Text != "邻接矩阵与邻接表" {
		t.Errorf("第二条要点 = %q", e.Points[1].Text)
	}
}

// TestEditorValidationShown 确认校验失败时弹窗不关、错误可见。
func TestEditorValidationShown(t *testing.T) {
	a := sampleApp(t)
	a.Router.Push("/study")
	tt := newTester(t, a, 1200, 820)

	if err := tt.Click("写学记"); err != nil {
		t.Fatal(err)
	}
	before := len(a.Day(a.Today).Entries)
	if err := tt.Click("保存"); err != nil {
		t.Fatal(err)
	}
	if !a.Editor.Open {
		t.Errorf("校验失败不该关弹窗")
	}
	if !strings.Contains(strings.Join(tt.Texts(), "\n"), "请填写科目") {
		t.Errorf("错误提示没有显示: %q", tt.Texts())
	}
	if len(a.Day(a.Today).Entries) != before {
		t.Errorf("不该写入任何记录")
	}
}

// TestCheckPointThroughUI 打勾一个要点，值要写进 model 并落库。
func TestCheckPointThroughUI(t *testing.T) {
	a := sampleApp(t)
	a.Router.Push("/study")
	tt := newTester(t, a, 1200, 820)
	tt.Advance(16 * 1000 * 1000)

	if err := tt.Click("图的基本概念"); err != nil {
		t.Fatalf("点要点: %v", err)
	}
	day := a.Day(a.Today)
	if !day.Entries[1].Points[0].Done {
		t.Fatalf("要点没有被打勾: %+v", day.Entries[1].Points)
	}
}

// TestTodoThroughUI 在待办页添加、勾选、删除。
func TestTodoThroughUI(t *testing.T) {
	a := sampleApp(t)
	a.Router.Push("/todo")
	tt := newTester(t, a, 1200, 820)

	before := len(a.Todos)
	if err := tt.Click("新待办"); err != nil {
		t.Fatalf("点输入框: %v", err)
	}
	tt.Type("复习 B 树")
	tt.Key(0, ui.KeyEnter)
	if len(a.Todos) != before+1 {
		t.Fatalf("回车后待办数 = %d, 期望 %d", len(a.Todos), before+1)
	}
	if a.TodoDraft != "" {
		t.Errorf("添加后输入框应清空")
	}
}

// TestThemeSwitchThroughUI 点侧边栏的外观按钮切换主题，并确认落库。
func TestThemeSwitchThroughUI(t *testing.T) {
	a := sampleApp(t)
	a.Settings.Theme = model.ThemeDark
	a.Settings.AutoBackup = false
	tt := newTester(t, a, 1200, 820)

	if err := tt.Click("切换到浅色"); err != nil {
		t.Fatalf("点外观按钮: %v", err)
	}
	if a.Settings.Theme != model.ThemeLight {
		t.Fatalf("Theme = %q, 期望 light", a.Settings.Theme)
	}
	st, err := a.Store.Settings()
	if err != nil {
		t.Fatal(err)
	}
	if st.Theme != model.ThemeLight {
		t.Errorf("外观没有落库: %+v", st)
	}
}

// TestSearchHighlightThroughUI 搜索后命中文字仍然完整显示（高亮不改变内容）。
func TestSearchHighlightThroughUI(t *testing.T) {
	a := sampleApp(t)
	a.Router.Push("/study")
	tt := newTester(t, a, 1200, 820)

	a.Query = "虚拟存储器"
	tt.Frame()
	if !tt.HasText("虚拟存储器的基本概念") {
		t.Errorf("高亮后文字应该原样可见: %q", tt.Texts())
	}
	if !tt.HasText("「虚拟存储器」· 2 天 · 3 条要点") {
		t.Errorf("副标题不对: %q", tt.Texts())
	}
}

// TestOnboardingThroughUI 检查首次启动的称呼引导。
func TestOnboardingThroughUI(t *testing.T) {
	a := sampleApp(t)
	a.Settings.Onboarded, a.Settings.UserName = false, ""
	a.Onboarding = app.OnboardingState{Open: true, Name: ""}
	tt := newTester(t, a, 1200, 820)

	if err := tt.Click("开始使用"); err != nil {
		t.Fatal(err)
	}
	if a.Onboarding.Err == "" {
		t.Errorf("空称呼应该提示错误")
	}

	tt.Type("小明")
	if err := tt.Click("开始使用"); err != nil {
		t.Fatal(err)
	}
	if a.Onboarding.Open {
		t.Errorf("填了称呼之后引导应该关闭")
	}
	if a.Settings.UserName != "小明" || !a.Settings.Onboarded {
		t.Errorf("设置 = %+v", a.Settings)
	}
}

// TestShortcuts 检查页面快捷键。
func TestShortcuts(t *testing.T) {
	a := sampleApp(t)
	tt := newTester(t, a, 1200, 820)

	tt.Key(ui.Cmd, ui.Key3)
	if got := a.Router.Path(); got != "/todo" {
		t.Errorf("⌘3 后路径 = %q, 期望 /todo", got)
	}
	tt.Key(ui.Cmd, ui.Key2)
	if got := a.Router.Path(); got != "/study" {
		t.Errorf("⌘2 后路径 = %q, 期望 /study", got)
	}
}

// TestShortcutsInertBehindModal 确认弹窗打开时，窗口级快捷键不会误触发。
func TestShortcutsInertBehindModal(t *testing.T) {
	a := sampleApp(t)
	a.Router.Push("/study")
	tt := newTester(t, a, 1200, 820)

	if err := tt.Click("写学记"); err != nil {
		t.Fatal(err)
	}
	tt.Type("数据结构")
	tt.Key(ui.Cmd, ui.Key3)
	if got := a.Router.Path(); got != "/study" {
		t.Errorf("⌘3 在弹窗里切走了页面: %q", got)
	}
	if a.Editor.Subject != "数据结构" {
		t.Errorf("⌘N 之类不应该清空草稿: %q", a.Editor.Subject)
	}
}
