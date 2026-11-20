package design

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// 字号阶梯（Material 3 的 type scale 裁剪到桌面端需要的几档）。
//
// 层次来自「字号 + 字重」：大标题 32/600，页面标题 26/600，区块标题 16/600，
// 正文 14/400，说明 12/400，极小号 11/500。不要再在控件上写裸数字。
const (
	// FontDisplay 是首页问候语。
	FontDisplay float32 = 32
	// FontTitle 是页面标题。
	FontTitle float32 = 26
	// FontHero 是寄语、统计数字。
	FontHero float32 = 24
	// FontHeading 是弹窗标题。
	FontHeading float32 = 20
	// FontLead 是区块标题、空状态标题。
	FontLead float32 = 16
	// FontBody 是正文：要点、待办、备注。
	FontBody float32 = 14
	// FontSmall 是次要正文：科目、日期、设置项的值。
	FontSmall float32 = 13
	// FontCaption 是说明文字、徽标、按钮里的小字。
	FontCaption float32 = 12
	// FontTiny 是极小号：日历、版本号、字段标签。
	FontTiny float32 = 11
)

// 圆角阶梯（Material 3 的 shape scale）。
const (
	// RadiusXS 是芯片、小标签。
	RadiusXS float32 = 8
	// RadiusSM 是输入框、图标按钮。
	RadiusSM float32 = 12
	// RadiusMD 是卡片。
	RadiusMD float32 = 20
	// RadiusLG 是弹窗、大的图片容器。
	RadiusLG float32 = 28
	// RadiusFull 是胶囊形按钮。
	RadiusFull float32 = 999
)

// Theme 把颜色板组装成 MyGo 的主题。
//
// MyGo 每帧都会用「跟随系统 + 桌面偏好」的默认主题重建上下文，
// 所以调用方每帧从 LightTheme/DarkTheme 出发，先套 WithPreferences（对比度、
// 字号缩放、减少动效），再用这里的角色覆盖色板。
//
// 参数 p 是颜色板，prefs 是桌面偏好，返回可直接 SetTheme 的主题。
func Theme(p Palette, prefs ui.Preferences) *ui.Theme {
	base := ui.LightTheme()
	if p.Dark {
		base = ui.DarkTheme()
	}
	t := *base.WithPreferences(prefs)

	t.Background = p.Background
	// 输入框、普通按钮的底色用高一层的容器，hover/pressed 再深一档。
	t.Surface = p.SurfaceContainerHigh
	t.SurfaceHover = p.SurfaceContainerHighest
	t.SurfacePressed = p.SurfaceContainer
	t.Border = p.OutlineVariant
	t.Text = p.OnSurface
	t.TextMuted = p.OnSurfaceVariant

	t = *t.WithAccent(p.Primary)
	// MD3 的主色是浅靛蓝时，白字对比度不够；主色文字统一用 OnPrimary。
	t.AccentText = p.OnPrimary
	t.AccentHover = p.Primary.Mix(p.OnPrimary, 0.12)
	t.AccentPressed = p.Primary.Mix(p.OnPrimary, 0.22)
	t.InactiveAccent = p.SurfaceContainerHighest

	t.Danger = p.Error
	t.Success = p.Tertiary
	t.Warning = ui.Hex("#B26A00")
	if p.Dark {
		t.Warning = ui.Hex("#FFB95C")
	}
	t.Selection = p.PrimaryContainer
	t.Focus = p.Primary
	t.Inverse = p.Inverse
	t.InverseText = p.InverseText
	t.Scrollbar = p.Outline
	t.ScrollbarWidth = 8

	t.Radius = RadiusSM
	// 4 是 MyGo 的默认间距单位：复选框方框 16、输入框内边距 6/10。
	t.Spacing = 4
	t.FontSize = FontBody
	t.TooltipDelay = 500 * time.Millisecond
	return &t
}
