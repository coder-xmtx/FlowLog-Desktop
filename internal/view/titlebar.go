package view

import (
	"flowlog-desktop/internal/app"
	design "flowlog-desktop/internal/design"

	"github.com/egoist/mygo/ui"
)

// titleBar 是自绘的窗口标题栏：整条可以拖动窗口，右上角是最小化 / 最大化 / 关闭。
//
// 窗口用 TitleBarHidden + HideWindowButtons 去掉了系统标题栏与按钮，
// 这条栏补上它们；双击空白处等于最大化（MyGo 的 DragWindow 自带）。
func titleBar(c *ui.Context, a *app.App) ui.Element {
	p := design.PaletteOf(c)

	bar := ui.Row(c).Height(38).FillWidth().AlignItems(ui.Center).Padding(0, 8).
		Background(p.SurfaceContainerLow).
		BorderWidth(0, 0, 1, 0).BorderColor(p.OutlineVariant)

	bar.Children(func() {
		// 左半边留给拖动：整块区域都是窗口的「把手」。
		ui.Row(c).Grow(1).FillHeight().DragWindow()

		windowButton(c, "win-min", "最小化", design.IconMinus, false, a.MinimizeWindow)
		icon := design.IconSquare
		tip := "最大化"
		if a.Maximized() {
			icon, tip = design.IconRestore, "还原"
		}
		windowButton(c, "win-max", tip, icon, false, a.ToggleMaximizeWindow)
		windowButton(c, "win-close", "关闭", design.IconClose, true, a.CloseWindow)
	})
	return bar
}

// windowButton 是标题栏上的一个窗口按钮。
//
// 参数 danger 为真时用错误色（关闭按钮），hover 是红色底。
func windowButton(c *ui.Context, key any, label string, icon *ui.SVG, danger bool, on func()) {
	p := design.PaletteOf(c)
	b := ui.ButtonBase(c.Key(key)).Label(label).Tooltip(label).
		Width(40).Height(30).Radius(design.RadiusXS).Center().Shrink(0)
	hovered := b.Hovered()

	fg := p.OnSurfaceVariant
	bg := p.OnSurface.Alpha(0)
	if hovered {
		bg = p.OnSurface.Alpha(0.08)
		fg = p.OnSurface
		if danger {
			bg = p.Error
			fg = p.OnError
		}
	}
	b.Children(func() {
		ui.Icon(c, icon).FontSize(design.FontLead).TextColor(fg)
	})
	b.Background(b.AnimateColor("bg", bg, design.MotionState))
	b.OnClick(on)
}
