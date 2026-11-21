package view

import (
	"time"

	"flowlog-desktop/internal/app"
	design "flowlog-desktop/internal/design"

	"github.com/egoist/mygo/ui"
)

// titleBarHeight 是自绘标题栏的高度。
const titleBarHeight float32 = 38

// titleBar 是自绘的窗口标题栏：左边是当前时间，整条可以拖动窗口，
// 右上角是最小化 / 最大化 / 关闭。
//
// 窗口用 TitleBarHidden + HideWindowButtons 去掉了系统标题栏与按钮，
// 这条栏补上它们；双击空白处等于最大化（MyGo 的 DragWindow 自带）。
//
// 它由 view.Root 放进 ui.Overlay 里画：弹窗的遮罩在普通层，标题栏在
// 弹层之上，所以开着「写学记」也能拖窗口。
func titleBar(c *ui.Context, a *app.App) ui.Element {
	p := design.PaletteOf(c)

	bar := ui.Row(c).Height(titleBarHeight).AlignItems(ui.Center).Padding(0, 8).
		Background(p.SurfaceContainerLow)

	bar.Children(func() {
		// 拖动手柄：整条减去右侧按钮的部分。
		handle := ui.Row(c).Grow(1).FillHeight().AlignItems(ui.Center).DragWindow()
		handle.Children(func() {
			clock(c)
		})

		windowButton(c, "win-min", design.T("最小化"), design.IconMinus, false, a.MinimizeWindow)
		icon := design.IconSquare
		tip := "最大化"
		if a.Maximized() {
			icon, tip = design.IconRestore, "还原"
		}
		windowButton(c, "win-max", design.T(tip), icon, false, a.ToggleMaximizeWindow)
		windowButton(c, "win-close", design.T("关闭"), design.IconClose, true, a.CloseWindow)
	})
	return bar
}

// clock 在标题栏左侧显示当前时间（24 小时制），每分钟重画一次。
func clock(c *ui.Context) {
	now := c.Now()
	c.After(time.Until(now.Truncate(time.Minute).Add(time.Minute)))

	p := design.PaletteOf(c)
	ui.Text(c, now.Format("15:04")).FontSize(design.FontSmall).
		TextColor(p.OnSurfaceVariant).Padding(0, 10, 0).SingleLine()
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
