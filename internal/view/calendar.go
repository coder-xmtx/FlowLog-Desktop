package view

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
)

// monthCalendar 是侧边栏里的月历：有日记的日子填色，今天描边。
//
// 点某一天会跳到日记页并滚到那一天，这是「翻旧账」最顺手的入口。
func (app *App) monthCalendar(c *ui.Context) {
	t := c.Theme()
	first := app.calMonth
	year, month := first.Year(), first.Month()

	ui.Column(c).Padding(8, 12, 6).Gap(6).Children(func() {
		ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
			navButton(c, "cal-prev", "上个月", "‹", func() {
				app.calMonth = first.AddDate(0, -1, 0)
			})

			title := ui.Text(c, fmt.Sprintf("%d 年 %d 月", year, int(month))).
				FontSize(11.5).FontWeight(600).TextColor(t.Text)
			title.Grow(1).TextAlign(ui.Center).Cursor(ui.CursorPointer).
				Tooltip("回到本月").Label("回到本月")
			if title.Clicked() {
				now := time.Now()
				app.calMonth = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
			}

			navButton(c, "cal-next", "下个月", "›", func() {
				app.calMonth = first.AddDate(0, 1, 0)
			})
		})

		// 星期表头，从周一开始
		ui.Row(c).Children(func() {
			for _, w := range []string{"一", "二", "三", "四", "五", "六", "日"} {
				ui.Box(c).Grow(1).AlignItems(ui.Center).Children(func() {
					ui.Text(c, w).FontSize(9.5).TextColor(t.TextMuted)
				})
			}
		})

		app.calendarGrid(c, first, int(month))
	})
}

// calendarGrid 画日期格子：先是月初的空格，再是 1..n。
func (app *App) calendarGrid(c *ui.Context, first time.Time, month int) {
	// 周一为一周的第一天：Go 的 Sunday=0，换算成 Monday=0。
	offset := (int(first.Weekday()) + 6) % 7
	days := time.Date(first.Year(), first.Month()+1, 0, 0, 0, 0, 0, time.Local).Day()
	rows := (offset + days + 6) / 7

	grid := ui.Grid(c).Columns(7).Gap(2)
	grid.Children(func() {
		for i := 0; i < rows*7; i++ {
			day := i - offset + 1
			if day < 1 || day > days {
				ui.Box(c).Height(24)
				continue
			}
			app.dayPill(c, first.Year(), month, day)
		}
	})
}

// dayPill 是月历里的一格。
func (app *App) dayPill(c *ui.Context, year, month, day int) {
	t := c.Theme()
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local).
		Format("2006-01-02")
	has := app.hasEntries(date)

	fill, fg := ui.Transparent, t.TextMuted
	if has {
		fill, fg = t.Accent, t.AccentText
	}

	label := fmt.Sprintf("%d 月 %d 日", month, day)
	if has {
		label += "，有日记"
	}

	box := ui.Box(c).Key("cal-" + date).Height(24).Radius(6).
		Background(fill).Center().Label(label).Tooltip(label).
		Cursor(ui.CursorPointer)
	if date == app.today {
		// 今天描边：有日记的格子用深色边，否则用强调色边，两种都能看见。
		ring := accentInk(t)
		if has {
			ring = t.AccentText
		}
		box.Border(1.5, ring)
	}
	box.Children(func() {
		txt := ui.Text(c, fmt.Sprintf("%d", day)).FontSize(11).TextColor(fg)
		if has {
			txt.FontWeight(600)
		}
	})
	box.OnClick(func() {
		app.revealDate = date
		app.router.Push("/diary")
	})
}

// navButton 是月历左右的小箭头。
func navButton(c *ui.Context, key, label, glyph string, on func()) {
	b := ui.Button(c.Key(key), "").Label(label).Tooltip(label).
		Size(20, 20).Padding(0).Radius(5).MinWidth(0)
	b.Children(func() {
		ui.Text(c, glyph).FontSize(12)
	})
	b.OnClick(on)
}
