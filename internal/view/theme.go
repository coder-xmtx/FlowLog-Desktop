// 这个文件集中放「视觉常量」：颜色、字号阶梯，以及主题的组装。
//
// 界面不用任何自定义字体：全部走系统默认字体（Windows 微软雅黑、
// macOS 苹方、Linux 由 fontconfig 决定），层次只靠字号与字重区分。
// 这样三端渲染一致，也不用为了一个字体往安装包里塞十几 MB。
package view

import (
	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// brandOrange 是流记的品牌色，取自 resources/icon.png 的波浪橙。
//
// 它很亮：白字在上面只有 2.4:1，所以按钮文字改成近黑色（8.8:1），
// 而「作为文字」的用法（链接、图标）在浅色主题下会再压暗一档，见 accentInk。
const brandOrange = "#ff8904"

// accentText 是放在品牌橙实底上的文字颜色。
const accentText = "#241205"

// 字号阶梯：界面里的文字只用这几档，不再随手写数字。
//
// 层次来自「字号 + 字重」：标题大而粗，正文中等，说明文字小而淡。
// 想整体放大或缩小界面，改这几个数字就够了。
const (
	// fsDisplay 是首页问候语，全界面最大的一行字。
	fsDisplay = 30
	// fsTitle 是页面标题（首页除外的四个页面）。
	fsTitle = 26
	// fsHero 是寄语、统计数字，以及关于页里的应用名。
	fsHero = 23
	// fsHeading 是弹窗标题与条目标题。
	fsHeading = 18
	// fsLead 是区块标题、空状态标题。
	fsLead = 16
	// fsBody 是正文：要点、待办、备注。
	fsBody = 14
	// fsSmall 是次要正文：科目、日期、设置项的值。
	fsSmall = 13
	// fsCaption 是说明文字、徽标、按钮里的小字。
	fsCaption = 12
	// fsTiny 是极小号：日历、版本号、字段标签。
	fsTiny = 11
)

// effectiveDark 报告当前应该用深色还是浅色：
// 深色 / 浅色是固定值，「跟随系统」读系统外观。
//
// 参数 c 是当前帧的上下文；它的默认主题已经跟随系统外观与桌面偏好。
func (app *App) effectiveDark(c *ui.Context) bool {
	switch app.settings.Theme {
	case model.ThemeLight:
		return false
	case model.ThemeDark:
		return true
	}
	return c.Theme().Dark
}

// applyTheme 把窗口主题换成流记的浅色或深色。
//
// MyGo 每帧都会用「跟随系统外观 + 桌面偏好」的默认主题重建上下文，
// 所以这里每帧从干净的 LightTheme/DarkTheme 出发：
// 先覆盖色板，再用 WithAccent 推导按钮、选中态等派生颜色。
func (app *App) applyTheme(c *ui.Context) {
	dark := app.effectiveDark(c)
	app.lastDark = dark

	base := ui.LightTheme()
	if dark {
		base = ui.DarkTheme()
	}
	t := *base.WithPreferences(c.Preferences())
	if dark {
		// 深色：偏暖的近黑。
		t.Background = ui.Hex("#17130f")
		t.Surface = ui.Hex("#211c17")
		t.SurfaceHover = ui.Hex("#2b251f")
		t.SurfacePressed = ui.Hex("#352e26")
		t.Border = ui.Hex("#3a322a")
		t.Text = ui.Hex("#f6f1ea")
		t.TextMuted = ui.Hex("#a89e92")
	} else {
		// 浅色：暖纸色，比纯白柔和，长时间看不刺眼。
		t.Background = ui.Hex("#f2ece3")
		t.Surface = ui.Hex("#e9e1d3")
		t.SurfaceHover = ui.Hex("#e0d6c5")
		t.SurfacePressed = ui.Hex("#d5c9b5")
		t.Border = ui.Hex("#d9cebc")
		t.Text = ui.Hex("#1f1a15")
		t.TextMuted = ui.Hex("#6e6459")
	}
	t.Radius = 8

	t = *t.WithAccent(ui.Hex(brandOrange))
	// 品牌橙够亮，不用白字：白字对比度只有 2.4，近黑有 8.8。
	t.AccentText = ui.Hex(accentText)
	c.SetTheme(&t)
}

// accentInk 返回强调色「作为文字 / 图标」时的颜色。
//
// 深色主题直接用品牌橙；浅色主题压暗，保证在浅底上的可读性。
// 参数 t 是当前主题。
func accentInk(t *ui.Theme) ui.Color {
	if t.Dark {
		return t.Accent
	}
	return t.Accent.Mix(ui.Hex("#000000"), 0.52)
}

// highlightBG 返回搜索命中片段的高亮底色。
//
// 浅色主题用一块低饱和的浅橙（像荧光笔），深色主题用品牌橙的低透明度，
// 两者都只改变底色，文字颜色仍由列表本身决定。
func highlightBG(t *ui.Theme) ui.Color {
	if t.Dark {
		return t.Accent.Alpha(0.28)
	}
	return ui.Hex("#ffcf8a").Alpha(0.75)
}

// toggleTheme 在浅色与深色之间切换，并写库。
//
// 「跟随系统」时按当前实际显示的那一套取反。
func (app *App) toggleTheme() {
	if app.effectiveDarkNow() {
		app.settings.Theme = model.ThemeLight
	} else {
		app.settings.Theme = model.ThemeDark
	}
	app.saveSettings()
}

// effectiveDarkNow 是给没有 Context 的调用点用的近似值：
// 「跟随系统」时读上一帧实际用的深浅色（app.lastDark）。
func (app *App) effectiveDarkNow() bool {
	switch app.settings.Theme {
	case model.ThemeLight:
		return false
	case model.ThemeDark:
		return true
	}
	return app.lastDark
}
