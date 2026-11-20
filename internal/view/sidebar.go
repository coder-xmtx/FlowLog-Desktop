package view

import (
	"strconv"

	"flowlog-desktop/internal/app"
	design "flowlog-desktop/internal/design"

	"github.com/egoist/mygo/ui"
)

// Sidebar 是左侧导航：品牌、四个页面、月历、外观与版本。
//
// 导航项是自己用 ButtonBase 拼的，为了做出 Material 3 的「药丸指示器」：
// 选中时主色容器底 + 主色文字，悬停时平滑浮现一层 state layer。
func Sidebar(c *ui.Context, a *app.App) {
	p := design.PaletteOf(c)
	side := ui.Column(c).Width(244).Shrink(0).FillHeight().
		Background(p.SurfaceContainerLow).Padding(12, 12, 12).Gap(4)

	side.Children(func() {
		brand(c)

		ui.Column(c).Gap(2).Children(func() {
			navItem(c, a, "/", design.IconHome, "首页", 0)
			navItem(c, a, "/study", design.IconBook, "学记", a.TodayEntryCount())
			navItem(c, a, "/notes", design.IconNote, "随笔", a.TodayNoteCount())
			navItem(c, a, "/todo", design.IconCheckCircle, "待办", a.OpenTodoCount())
			navItem(c, a, "/settings", design.IconSettings, "设置", 0)
		})

		ui.Spacer(c)
		MonthCalendar(c, a)
		sidebarFooter(c, a)
	})
}

// brand 是侧边栏顶部的标志：应用图标 + 名称。
func brand(c *ui.Context) {
	p := design.PaletteOf(c)
	ui.Row(c).Gap(10).Padding(8, 10, 18).AlignItems(ui.Center).Children(func() {
		if bmp := app.Icon(); bmp != nil {
			ui.Image(c, bmp).Size(32, 32).Radius(0).Shrink(0)
		} else {
			ui.Box(c).Size(32, 32).Radius(5).Background(p.Primary).Shrink(0)
		}
		ui.Column(c).Gap(1).Children(func() {
			ui.Text(c, app.Name).FontSize(17).FontWeight(600).TextColor(p.OnSurface).SingleLine()
			ui.Text(c, "FLOWLOG").FontSize(design.FontTiny).FontWeight(500).
				LetterSpacing(1.6).TextColor(p.OnSurfaceVariant).SingleLine()
		})
	})
}

// navItem 是一个导航项。
//
// 参数 path 是路由路径，icon 是图标，label 是文字，badge 大于 0 时显示数量。
func navItem(c *ui.Context, a *app.App, path string, icon *ui.SVG, label string, badge int) {
	p := design.PaletteOf(c)
	active := a.Router.Path() == path

	item := ui.ButtonBase(c.Key("nav-"+path)).Height(44).Padding(0, 14).
		Radius(design.RadiusFull).Gap(12).Shrink(0).Label(label)
	hovered := item.Hovered()

	fill, fg := ui.Transparent, p.OnSurfaceVariant
	if active {
		fill, fg = p.PrimaryContainer, p.OnPrimaryContainer
	}

	item.Children(func() {
		ui.Icon(c, icon).FontSize(design.FontHeading).TextColor(fg)
		ui.Text(c, label).FontSize(design.FontBody).FontWeight(500).
			TextColor(fg).Grow(1).SingleLine().Bold()
		if badge > 0 {
			design.Chip(c, strconv.Itoa(badge), design.ToneSecondary)
		}
	})

	// 背景动画：没选中时从透明渐入一层 state layer，选中时直接是药丸。
	target := fill
	if !active {
		target = design.HoverLayer(item, "hover", p.OnSurface, hovered)
	}
	item.Background(item.AnimateColor("bg", target, design.MotionState))
	item.OnClick(func() { a.Router.Push(path) })
}

// sidebarFooter 是侧边栏底部：外观切换与版本号。
func sidebarFooter(c *ui.Context, a *app.App) {
	p := design.PaletteOf(c)
	ui.Row(c).Gap(6).AlignItems(ui.Center).Padding(6, 2, 2).Children(func() {
		icon := design.IconSun
		label := "浅色"
		tip := "切换到浅色"
		if !a.LastDark {
			icon, label, tip = design.IconMoon, "深色", "切换到深色"
		}
		design.IconButton(c, "theme-toggle", tip, icon, a.ToggleTheme)
		ui.Text(c, label).FontSize(design.FontCaption).TextColor(p.OnSurfaceVariant).
			Tooltip(tip).SingleLine()
		ui.Spacer(c)
		ui.Text(c, "v"+app.Version).FontSize(design.FontTiny).TextColor(p.OnSurfaceVariant)
	})
}
