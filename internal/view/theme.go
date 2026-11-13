package view

import (
	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// brandOrange 是流记的品牌色，取自 resources/icon.png 的波浪橙。
//
// 界面只保留这一个强调色：白字在 #c2410c 上的对比度约 5.2:1，
// 13px 的按钮文字也清楚。
const brandOrange = "#c2410c"

// fontSerif 是标题用的衬线字体栈：拉丁字母用 Georgia，
// 中文交给 Noto Serif SC / 思源宋体 / 宋体，缺字时逐级回退。
const fontSerif = "Georgia, Cambria, Times New Roman, Noto Serif SC, Source Han Serif CN, STSong, SimSun, Songti SC"

// applyTheme 把窗口主题换成流记的浅色或深色。
//
// MyGo 每帧都会用「跟随系统外观 + 桌面偏好」的默认主题重建上下文，
// 所以这里每帧从干净的 LightTheme/DarkTheme 出发：
//  1. WithPreferences 先吃下系统的对比度与字号缩放；
//  2. 再覆盖成暖纸色的中性色板；
//  3. 最后 WithAccent 固定品牌橙。
func (app *App) applyTheme(c *ui.Context) {
	dark := app.settings.Theme == model.ThemeDark

	base := ui.LightTheme()
	if dark {
		base = ui.DarkTheme()
	}
	t := *base.WithPreferences(c.Preferences())
	if dark {
		// 深色：偏暖的近黑，而不是纯灰的冷调。
		t.Background = ui.Hex("#17130f")
		t.Surface = ui.Hex("#211c17")
		t.SurfaceHover = ui.Hex("#2b251f")
		t.SurfacePressed = ui.Hex("#352e26")
		t.Border = ui.Hex("#3a322a")
		t.Text = ui.Hex("#f6f1ea")
		t.TextMuted = ui.Hex("#a89e92")
	} else {
		// 浅色：像纸一样的暖白。
		t.Background = ui.Hex("#fffcf8")
		t.Surface = ui.Hex("#f4eee6")
		t.SurfaceHover = ui.Hex("#ece4d9")
		t.SurfacePressed = ui.Hex("#e3d9cb")
		t.Border = ui.Hex("#e6ddd0")
		t.Text = ui.Hex("#1c1917")
		t.TextMuted = ui.Hex("#79706a")
	}
	t.Radius = 8

	t = *t.WithAccent(ui.Hex(brandOrange))
	c.SetTheme(&t)
}

// toggleTheme 在浅色与深色之间切换，并写库。
func (app *App) toggleTheme() {
	if app.settings.Theme == model.ThemeDark {
		app.settings.Theme = model.ThemeLight
	} else {
		app.settings.Theme = model.ThemeDark
	}
	app.saveSettings()
}
