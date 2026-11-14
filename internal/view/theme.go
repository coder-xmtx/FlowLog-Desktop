package view

import (
	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// brandOrange 是流记的品牌色，取自 resources/icon.png 的波浪橙。
//
// 它很亮：白字在上面只有 2.4:1，所以按钮文字改成近黑色（8.8:1），
// 而"作为文字"的用法（链接、图标）在浅色主题下会再压暗一档，见 accentInk。
const brandOrange = "#ff8904"

// accentText 是放在品牌橙实底上的文字颜色。
const accentText = "#241205"

// fontSerif 是标题用的衬线字体栈：拉丁字母用 Georgia，
// 中文交给 Noto Serif SC / 思源宋体 / 宋体，缺字时逐级回退。
const fontSerif = "Georgia, Cambria, Times New Roman, Noto Serif SC, Source Han Serif CN, STSong, SimSun, Songti SC"

// effectiveDark 报告当前应该用深色还是浅色：
// 深色 / 浅色是固定值，"跟随系统"读系统外观。
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
// 所以这里每帧从干净的 LightTheme/DarkTheme 出发。
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

// accentInk 返回强调色"作为文字/图标"时的颜色。
//
// 深色主题直接用品牌橙；浅色主题压暗，保证在浅底上的可读性。
func accentInk(t *ui.Theme) ui.Color {
	if t.Dark {
		return t.Accent
	}
	return t.Accent.Mix(ui.Hex("#000000"), 0.52)
}

// toggleTheme 在浅色与深色之间切换，并写库。
func (app *App) toggleTheme() {
	if app.effectiveDarkNow() {
		app.settings.Theme = model.ThemeLight
	} else {
		app.settings.Theme = model.ThemeDark
	}
	app.saveSettings()
}

// effectiveDarkNow 是给没有 Context 的调用点用的近似值：
// "跟随系统"时按上次主题绘制的结果无法得知，这里只处理固定值。
func (app *App) effectiveDarkNow() bool {
	switch app.settings.Theme {
	case model.ThemeLight:
		return false
	case model.ThemeDark:
		return true
	}
	return app.lastDark
}
