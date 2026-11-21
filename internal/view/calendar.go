package view

import (
	"strconv"
	"time"

	"flowlog-desktop/internal/app"
	design "flowlog-desktop/internal/design"
	"flowlog-desktop/internal/lang"
	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// MonthCalendar 是侧边栏里的月历：有记录的日子填主色容器色，今天描边。
//
// 点某一天会跳到学记页并滚到那一天，这是「翻旧账」最顺手的入口。
func MonthCalendar(c *ui.Context, a *app.App) {
	p := design.PaletteOf(c)
	first := a.CalMonth
	year, month := first.Year(), first.Month()

	ui.Column(c).Padding(4, 6, 6).Gap(8).Children(func() {
		ui.Row(c).Gap(2).AlignItems(ui.Center).Children(func() {
			design.IconButton(c, "cal-prev", design.T("上个月"), design.IconLeft, func() {
				a.CalMonth = first.AddDate(0, -1, 0)
			})

			title := design.Text(c, model.MonthTitle(year, month)).
				FontSize(design.FontCaption).FontWeight(600).TextColor(p.OnSurface)
			title.Grow(1).TextAlign(ui.Center).Cursor(ui.CursorPointer).
				Tooltip(design.T("回到本月")).Label(design.T("回到本月"))
			if title.Clicked() {
				now := time.Now()
				a.CalMonth = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
			}

			design.IconButton(c, "cal-next", design.T("下个月"), design.IconRight, func() {
				a.CalMonth = first.AddDate(0, 1, 0)
			})
		})

		// 星期表头，从周一开始
		ui.Row(c).Children(func() {
			for i := 0; i < 7; i++ {
				w := model.WeekdayHead(time.Date(2024, 1, 1+i, 0, 0, 0, 0, time.Local))
				ui.Box(c).Grow(1).AlignItems(ui.Center).Children(func() {
					design.Text(c, w).FontSize(design.FontTiny).TextColor(p.OnSurfaceVariant)
				})
			}
		})

		calendarGrid(c, a, first, int(month))
	})
}

// calendarGrid 画日期格子：先是月初的空格，再是 1..n。
func calendarGrid(c *ui.Context, a *app.App, first time.Time, month int) {
	// 周一为一周的第一天：Go 的 Sunday=0，换算成 Monday=0。
	offset := (int(first.Weekday()) + 6) % 7
	days := time.Date(first.Year(), first.Month()+1, 0, 0, 0, 0, 0, time.Local).Day()
	rows := (offset + days + 6) / 7

	grid := ui.Grid(c).Columns(7).Gap(2)
	grid.Children(func() {
		for i := 0; i < rows*7; i++ {
			day := i - offset + 1
			if day < 1 || day > days {
				ui.Box(c).Height(28)
				continue
			}
			dayPill(c, a, first.Year(), month, day)
		}
	})
}

// dayPill 是月历里的一格。
func dayPill(c *ui.Context, a *app.App, year, month, day int) {
	p := design.PaletteOf(c)
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local).
		Format("2006-01-02")
	has := a.HasEntries(date)
	today := date == a.Today

	fill, fg := ui.Transparent, p.OnSurfaceVariant
	if has {
		fill, fg = p.PrimaryContainer, p.OnPrimaryContainer
	}

	label := model.DayLabel(year, time.Month(month), day)
	if has {
		label += lang.T("，有记录")
	}

	box := ui.Box(c.Key("cal-" + date)).Height(28).Radius(14).Center().
		Label(label).Tooltip(label).Cursor(ui.CursorPointer)
	if today {
		ring := p.Primary
		if has {
			ring = p.OnPrimaryContainer
		}
		box.Border(1.5, ring)
	}
	box.Children(func() {
		txt := ui.Text(c, strconv.Itoa(day)).FontSize(design.FontCaption).TextColor(fg)
		if has {
			txt.FontWeight(600)
		}
	})
	// 悬停时浮起一层很淡的主色，鼠标移上去有过渡而不是突然变化。
	target := fill
	if box.Hovered() {
		target = fill.Mix(p.Primary, 0.10)
	}
	box.Background(box.AnimateColor("hover", target, design.MotionState))
	box.OnClick(func() {
		a.RevealDate = date
		a.Router.Push("/study")
	})
}
