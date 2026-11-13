package view

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// quotes 是首页每天轮换的寄语，按「一年中的第几天」取，同一天固定。
//
// 都是随手的句子，想换成自己的话直接改这个切片即可。
var quotes = []string{
	"把今天弄懂的一个点，写成一句话，它就不会在明天跑掉。",
	"进度不是一天走了多远，而是今天有没有往前走一点。",
	"看不懂的地方，往往只是还没拆成足够小的问题。",
	"复习不是重读，而是合上书，把记得的写下来。",
	"一天里最值钱的十分钟，是开始的那十分钟。",
	"记住一个概念的最好方式，是把它讲给别人听。",
	"别急着学新的，先把昨天的漏洞补上。",
	"慢一点没关系，别停。",
	"真正学会的东西，会在很久以后还认得你。",
	"把难的那一门放在状态最好的时候。",
	"写下「我卡在哪」，问题就解决了一半。",
	"今天的笔记，是明天的自己需要的说明书。",
	"所有看起来很快的人，都有一段没人看见的慢。",
	"会做的题定义了你现在的水平，不会的题决定了你要去哪儿。",
	"输入让人安心，输出才让人进步。",
	"把大目标切到今天能做完的那一小块。",
	"困惑是理解的开始，不是失败的证据。",
	"每天留一点时间给「想不明白但很重要」的问题。",
	"记忆靠重复，理解靠连接。",
	"今天做过的事，写下来才算真的发生过。",
	"与其计划得很完美，不如现在写第一行。",
	"把不会的标出来，比把会的再做一遍有用。",
	"学习是复利，今天看不到变化，半年后看得见。",
	"别和进度条较劲，和自己昨天的理解较劲。",
}

// quoteOf 返回某个日期的寄语。
func quoteOf(date string) string {
	t := model.MustDate(date)
	if t.IsZero() || len(quotes) == 0 {
		return quotes[0]
	}
	return quotes[model.DayNumber(t)%len(quotes)]
}

// streak 计算连续记录的天数。
//
// 今天还没写不会把连续记录清零：从昨天往前数，昨天写过就还算连续。
// 这样用户早上打开应用时看到的是「连续 5 天」，而不是「0 天」。
func (app *App) streak() int {
	day := model.MustDate(app.today)
	if day.IsZero() {
		return 0
	}
	if i := app.dayIndex(app.today); i < 0 || !app.days[i].HasContent() {
		day = day.AddDate(0, 0, -1)
	}
	n := 0
	// 上限只是防御：真有十年不断更的日记，也没必要每帧数下去。
	for ; n < 3650; n++ {
		i := app.dayIndex(day.Format(model.DateLayout))
		if i < 0 || !app.days[i].HasContent() {
			break
		}
		day = day.AddDate(0, 0, -1)
	}
	return n
}

// greeting 按当前时间返回问候语。
func greeting(now int) string {
	switch {
	case now < 5:
		return "夜深了"
	case now < 9:
		return "早上好"
	case now < 11:
		return "上午好"
	case now < 13:
		return "中午好"
	case now < 18:
		return "下午好"
	case now < 23:
		return "晚上好"
	}
	return "夜深了"
}

// home 是启动页：先给一点情绪价值，再给一眼能看完的今天。
func (app *App) home(c *ui.Context) {
	t := c.Theme()
	date := model.MustDate(app.today)

	today := app.day(app.today)
	streak := app.streak()
	totalDays := app.daysWritten()
	points := today.PointCount()

	ui.Column(c).Grow(1).MinWidth(0).Children(func() {
		ui.Column(c).Padding(30, 32, 10).Gap(5).Children(func() {
			ui.Text(c, greeting(c.Now().Hour())+"，").FontSize(27).Bold().SingleLine()
			ui.Text(c, model.DisplayDateFull(app.today)+" · "+model.Weekday(date)).
				FontSize(12.5).TextColor(t.TextMuted).SingleLine()
		})

		ui.Scroll(c).Grow(1).Padding(8, 32, 28).Gap(16).Children(func() {
			app.hero(c, quoteOf(app.today))

			ui.Row(c).Gap(14).Children(func() {
				app.statCard(c, iconSparkle, "连续记录", plural(streak, "天"), "不断更的记录")
				app.statCard(c, iconBook, "累计记录", plural(totalDays, "天"), "写下过的日子")
				app.statCard(c, iconCheckCircle, "今日要点", plural(points, "条"), "今天记下的知识点")
			})

			ui.Row(c).Gap(16).AlignItems(ui.Stretch).Children(func() {
				ui.Column(c).Grow(1.3).MinWidth(0).Children(func() {
					app.todayLearning(c, today).FillHeight()
				})
				ui.Column(c).Grow(1).MinWidth(0).Children(func() {
					app.todayTodos(c).FillHeight()
				})
			})

			app.recentStrip(c)
		})
	})
}

// hero 是首页顶部那张渐变卡片，三道波浪呼应应用图标。
func (app *App) hero(c *ui.Context, quote string) {
	t := c.Theme()
	// 从强调色的亮端渐变到暗端。两头都比纯色更耐看，
	// 白字在亮的那一头也还有 4.4 左右的对比度（白字在 #ff8904 上只有 2.4）。
	from := t.Accent.Mix(ui.Hex("#ffffff"), 0.14)
	to := t.Accent.Mix(ui.Hex("#000000"), 0.32)

	box := ui.Box(c).Height(180).Radius(18).Gradient(from, to, 115).Clip()
	box.Draw(func(p *ui.Painter, r ui.Rect) {
		drawWaves(p, r, ui.RGBA(255, 255, 255, 0.09), 11, 2)
	})
	box.Children(func() {
		ui.Column(c).Fill().Justify(ui.Center).Padding(26, 28).Gap(12).Children(func() {
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, iconSparkle).FontSize(13).TextColor(ui.RGBA(255, 255, 255, 0.88))
				ui.Text(c, "今日寄语").FontSize(11).FontWeight(600).LetterSpacing(1.2).
					TextColor(ui.RGBA(255, 255, 255, 0.88))
			})
			ui.Text(c, quote).FontSize(20.5).FontWeight(600).LineHeight(1.5).
				TextColor(ui.RGBA(255, 255, 255, 0.97)).MaxWidth(660).Balance()
		})
	})
}

// statCard 是一张小统计卡。
func (app *App) statCard(c *ui.Context, icon *ui.SVG, label, value, hint string) {
	t := c.Theme()
	ui.Column(c).Grow(1).Padding(16).Gap(8).Radius(14).
		Background(t.Background).Border(1, t.Border).
		Shadow(0, 1, 2, 0, ui.RGBA(0, 0, 0, 0.04)).
		Children(func() {
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, icon).FontSize(13).TextColor(t.Accent)
				ui.Text(c, label).FontSize(11.5).TextColor(t.TextMuted).SingleLine()
			})
			ui.Text(c, value).FontSize(21).Bold().SingleLine()
			ui.Text(c, hint).FontSize(11).TextColor(t.TextMuted).SingleLine()
		})
}

// todayLearning 首页左栏：今天记下的学习条目。
func (app *App) todayLearning(c *ui.Context, today model.Day) ui.Element {
	t := c.Theme()
	const max = 3

	return sectionCard(c, "今天学了什么", "", func() {
		if len(today.Entries) == 0 {
			emptyState(c, iconBook, "今天还没有记录",
				"写下今天弄懂的三个点，明天会感谢现在的你。",
				"写今天的日记", func() { app.openEditor(app.today, "") })
			return
		}

		for i, e := range today.Entries {
			if i == max {
				break
			}
			ui.Column(c).Gap(5).Children(func() {
				ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
					dot(c, 6, t.Accent)
					ui.Text(c, e.Subject).FontSize(13).FontWeight(600).SingleLine()
					if parts := model.TopicParts(e.Topic); len(parts) > 0 {
						ui.Text(c, strings.Join(parts, " › ")).FontSize(11.5).
							TextColor(t.TextMuted).SingleLine().Grow(1)
					}
				})
				// 有要点就露第一条，只有备注就露备注：让卡片永远有话可说。
				switch {
				case len(e.Points) > 0:
					ui.Text(c, "· "+e.Points[0].Text).FontSize(12.5).
						TextColor(t.Text).SingleLine().Padding(0, 14).Opacity(0.85)
				case strings.TrimSpace(e.Note) != "":
					ui.Text(c, "· "+firstLine(e.Note)).FontSize(12.5).
						TextColor(t.Text).SingleLine().Padding(0, 14).Opacity(0.85)
				}
			})
		}

		if extra := len(today.Entries) - max; extra > 0 {
			ui.Textf(c, "还有 %d 条记录", extra).FontSize(11.5).TextColor(t.TextMuted)
		}

		ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
			ui.Link(c, "在日记里看全部", "/diary").FontSize(12)
		})
	})
}

// todayTodos 首页右栏：今天的待办，可以直接勾。
func (app *App) todayTodos(c *ui.Context) ui.Element {
	t := c.Theme()
	today := app.day(app.today)
	done, total := today.TodoProgress()

	return sectionCard(c, "今日待办", "", func() {
		if total == 0 {
			ui.Text(c, "今天还没有待办。").FontSize(12.5).TextColor(t.TextMuted)
			ui.Button(c, "去添加").OnClick(func() { app.router.Push("/todo") })
			return
		}

		shown := 0
		for _, todo := range today.Todos {
			if todo.Done || shown == 4 {
				continue
			}
			app.todoCheckbox(c, todo.ID)
			shown++
		}
		if shown == 0 {
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, iconCheckCircle).FontSize(14).TextColor(t.Success)
				ui.Text(c, "今天的待办都做完了。").FontSize(12.5).TextColor(t.TextMuted)
			})
		}

		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Textf(c, "%d / %d 已完成", done, total).FontSize(11.5).TextColor(t.TextMuted)
			ui.Spacer(c)
			ui.Link(c, "全部待办", "/todo").FontSize(12)
		})
	})
}

// recentStrip 是首页底部的「最近 7 天」。
func (app *App) recentStrip(c *ui.Context) {
	ui.Column(c).Gap(10).Children(func() {
		ui.Text(c, "最近 7 天").FontSize(13.5).Bold()
		ui.Row(c).Gap(8).Children(func() {
			for i := 6; i >= 0; i-- {
				date := model.MustDate(app.today).AddDate(0, 0, -i)
				app.dayCell(c, date, i == 0)
			}
		})
	})
}

// dayCell 是最近 7 天里的一格：有记录的日子填上深一点的强调色。
//
// 用「强调色压暗 + 白字」而不是直接用 Accent + AccentText：
// 橙色的 AccentText 会自动选成近黑色，放在 12px 的数字上不够清楚。
func (app *App) dayCell(c *ui.Context, date time.Time, isToday bool) {
	t := c.Theme()
	key := date.Format(model.DateLayout)
	day := app.day(key)

	fill, fg := t.Surface, t.TextMuted
	if day.HasContent() {
		fill = t.Accent
		fg = t.AccentText
	}
	if isToday && !day.HasContent() {
		fg = t.Text
	}

	cell := ui.Column(c).Key("day-" + key).Width(46).Gap(6).AlignItems(ui.Center).Children(func() {
		ui.Text(c, model.WeekdayShort(date)).FontSize(10.5).TextColor(t.TextMuted)
		box := ui.Box(c).Size(32, 32).Radius(9).Background(fill)
		if isToday {
			box.Border(1.5, t.Accent)
		}
		box.Children(func() {
			ui.Text(c, strconv.Itoa(date.Day())).FontSize(12).FontWeight(600).
				TextColor(fg).Fill().Center()
		})
	})
	cell.Tooltip(key + " · " + daySummary(day))
}

// daySummary 用一句话概括某天，用于提示气泡。
func daySummary(day model.Day) string {
	if !day.HasContent() {
		return "没有记录"
	}
	parts := make([]string, 0, 2)
	if n := len(day.Entries); n > 0 {
		parts = append(parts, plural(n, "条学习记录"))
	}
	if done, total := day.TodoProgress(); total > 0 {
		parts = append(parts, fmt.Sprintf("待办 %d/%d", done, total))
	}
	return strings.Join(parts, " · ")
}
