// Package view 是流记的窗口外壳：根视图、侧边栏、月历与全局弹层。
//
// 页面在 internal/pages，状态与动作在 internal/app，样式在 internal/design。
// 这里只负责把它们拼成一个窗口，并处理窗口级的快捷键与主题。
package view

import (
	"time"

	"flowlog-desktop/internal/app"
	design "flowlog-desktop/internal/design"
	"flowlog-desktop/internal/model"
	"flowlog-desktop/internal/pages"

	"github.com/egoist/mygo/ui"
)

// Root 是窗口的内容函数：每帧被调用一次（或多次），把状态画成界面。
//
// 参数 a 是应用状态；main 里用 ui.View(func(c *ui.Context) { view.Root(c, a) }) 注册。
func Root(c *ui.Context, a *app.App) {
	a.RollOver()
	applyTheme(c, a)
	flushToasts(c, a)
	shortcuts(c, a)

	root := ui.Column(c).Fill()
	root.Children(func() {
		titleBar(c, a)
		ui.Row(c).Grow(1).AlignItems(ui.Stretch).Children(func() {
			Sidebar(c, a)
			a.Router.View(c, func(r *ui.Route) {
				switch {
				case r.Match("/"):
					r.Title("流记 · 首页")
					pages.Home(c, a)
				case r.Match("/study"):
					r.Title("流记 · 学记")
					pages.Study(c, a)
				case r.Match("/notes"):
					r.Title("流记 · 随笔")
					pages.Notes(c, a)
				case r.Match("/todo"):
					r.Title("流记 · 待办")
					pages.Todo(c, a)
				case r.Match("/settings"):
					r.Title("流记 · 设置")
					pages.Settings(c, a)
				default:
					ui.Text(c, "没有这个页面")
				}
			})
		})
	})

	// 弹层写在路由外面：切换页面时它们不会跟着消失。
	pages.StudyEditor(c, a)
	pages.NoteEditor(c, a)
	ConfirmDialog(c, a)
	OnboardingDialog(c, a)
}

// applyTheme 把窗口主题换成流记的浅色或深色。
//
// MyGo 每帧都会用「跟随系统外观 + 桌面偏好」的默认主题重建上下文，
// 所以这里每帧从干净的主题出发，再套上 design 的颜色板。
func applyTheme(c *ui.Context, a *app.App) {
	dark := effectiveDark(c, a)
	a.LastDark = dark
	c.SetTheme(design.Theme(design.For(dark), c.Preferences()))
}

// effectiveDark 报告当前应该用深色还是浅色：
// 深色 / 浅色是固定值，「跟随系统」读系统外观。
func effectiveDark(c *ui.Context, a *app.App) bool {
	switch a.Settings.Theme {
	case model.ThemeLight:
		return false
	case model.ThemeDark:
		return true
	}
	return c.Theme().Dark
}

// flushToasts 在帧的开头把攒下的提示交给窗口。
func flushToasts(c *ui.Context, a *app.App) {
	for _, t := range a.TakeToasts() {
		c.AddToast(ui.Toast{
			ID:          t.Title, // 同一个标题的提示就地更新，而不是叠一摞
			Title:       t.Title,
			Description: t.Desc,
			Type:        t.Type,
			Action:      t.Action,
			OnAction:    t.OnAct,
			Timeout:     4 * time.Second,
		})
	}
}

// shortcuts 处理窗口级快捷键。弹窗开着时 c.Shortcut 自动失效。
func shortcuts(c *ui.Context, a *app.App) {
	switch {
	case c.Shortcut(ui.Cmd, ui.KeyN):
		a.OpenEditor(a.Today, "")
	case c.Shortcut(ui.Cmd, ui.Key1):
		a.Router.Push("/")
	case c.Shortcut(ui.Cmd, ui.Key2):
		a.Router.Push("/study")
	case c.Shortcut(ui.Cmd, ui.Key3):
		a.Router.Push("/todo")
	case c.Shortcut(ui.Cmd, ui.Key4):
		a.Router.Push("/settings")
	case c.Shortcut(ui.Cmd, ui.KeyF):
		a.Router.Push("/study")
		a.FocusSearch = true
	}
}
