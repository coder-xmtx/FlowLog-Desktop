package view

import (
	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// brandAccent 是流记的品牌色，取自 resources/icon.png 的波浪橙。
//
// 图标用的是更亮的 #ff8904，但界面上的强调色要压暗一档：
// 白字放在 #ff8904 或 #ea580c 上的对比度只有 2.4~3.6，
// #c2410c 能到 5.2，13px 的按钮文字也看得清。
const brandAccent = "#c2410c"

// brandOrange 是图标上的亮橙，只用于 logo 这类装饰，不用于文字。
const brandOrange = "#ff8904"

// accentChoice 是一个可选的强调色。
type accentChoice struct {
	Name string
	Hex  string
}

// accentChoices 是「外观」面板里提供的强调色。
//
// 按钮上的文字颜色（AccentText）由 MyGo 按对比度自动选黑或白，
// 这里只负责挑颜色本身。
var accentChoices = []accentChoice{
	{Name: "流橙", Hex: brandAccent},
	{Name: "湖青", Hex: "#0f766e"},
	{Name: "靛蓝", Hex: "#4338ca"},
	{Name: "玫红", Hex: "#be123c"},
	{Name: "松绿", Hex: "#15803d"},
	{Name: "石墨", Hex: "#52525b"},
}

// accentIndexOf 返回强调色在 accentChoices 里的下标，未知值回到品牌橙。
func accentIndexOf(hex string) int {
	for i, a := range accentChoices {
		if a.Hex == hex {
			return i
		}
	}
	return 0
}

// accentColor 返回当前选中的强调色。
func (app *App) accentColor() ui.Color {
	if app.accentPick < 0 || app.accentPick >= len(accentChoices) {
		return ui.Hex(brandAccent)
	}
	return ui.Hex(accentChoices[app.accentPick].Hex)
}

// applyTheme 把窗口主题换成流记自己的那套。
//
// MyGo 每帧都会用「跟随系统外观 + 桌面偏好」的默认主题重建上下文，
// 所以这里每帧都从干净的 LightTheme/DarkTheme 出发是安全的：
//  1. WithPreferences 先吃下系统的强调色、对比度、字号缩放；
//  2. 再覆盖成暖纸色的中性色板；
//  3. 最后 WithAccent 换成流记橙（或用户选的色），
//     AccentHover / AccentPressed / Selection / Focus 会一起被推导出来。
//
// 「跟随系统」是真的跟随：用户在系统里切深浅色，下一帧就变。
func (app *App) applyTheme(c *ui.Context) {
	dark := c.Theme().Dark
	switch app.themePick {
	case 1: // 浅色
		dark = false
	case 2: // 深色
		dark = true
	}

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

	t = *t.WithAccent(app.accentColor())
	c.SetTheme(&t)
}

// themeModeName 返回外观模式的中文名。
func themeModeName(mode string) string {
	switch mode {
	case model.ThemeLight:
		return "浅色"
	case model.ThemeDark:
		return "深色"
	}
	return "跟随系统"
}

// themeModeIndex / themeModeAt 在三种外观之间换算。
func themeModeIndex(mode string) int {
	switch mode {
	case model.ThemeLight:
		return 1
	case model.ThemeDark:
		return 2
	}
	return 0
}

func themeModeAt(i int) string {
	switch i {
	case 1:
		return model.ThemeLight
	case 2:
		return model.ThemeDark
	}
	return model.ThemeSystem
}
