package view

import (
	"strconv"
	"strings"
	"time"

	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// quotes 是首页每天轮换的寄语，按「一年中的第几天」取，同一天固定。
var quotes = []string{
	"把今天弄懂的一点，写成一句话。",
	"进度的意义在于持续，而不在于快慢。",
	"看不懂的地方，通常只是还没有拆成足够小的问题。",
	"复习不是重读，而是合上书，把记得的写下来。",
	"一天里最有价值的十分钟，是开始的那十分钟。",
	"检验理解的方式，是把它讲给别人听。",
	"先把昨天的漏洞补上，再学新的。",
	"慢一点没有关系，不要停。",
	"真正学会的东西，会在很久以后仍然认得你。",
	"把最难的一门，放在状态最好的时候。",
	"写下「我卡在哪里」，问题就解决了一半。",
	"今天的笔记，是明天的自己的说明书。",
	"会做的题定义现在的水平，不会的题决定要去的地方。",
	"输入让人安心，输出才让人进步。",
	"把大目标切到今天能做完的那一块。",
	"困惑是理解的开始。",
	"记忆靠重复，理解靠连接。",
	"今天做过的事，写下来才算真的发生过。",
	"与其计划得完美，不如现在写第一行。",
	"学习是复利：今天看不到变化，半年后看得见。",
}

// quoteOf 返回某个日期的寄语。
func quoteOf(date string) string {
	t := model.MustDate(date)
	if t.IsZero() || len(quotes) == 0 {
		return quotes[0]
	}
	return quotes[model.DayNumber(t)%len(quotes)]
}

// greetingByHour 按小时返回问候语。
func greetingByHour(hour int) string {
	switch {
	case hour < 5:
		return "夜深了"
	case hour < 9:
		return "早上好"
	case hour < 11:
		return "上午好"
	case hour < 13:
		return "中午好"
	case hour < 18:
		return "下午好"
	case hour < 23:
		return "晚上好"
	}
	return "夜深了"
}

// home 是启动页：一句寄语、今天的概览、近期要做的待办。
func (app *App) home(c *ui.Context) {
	t := c.Theme()
	date := model.MustDate(app.today)

	today := app.day(app.today)
	streak := app.streak()
	totalDays := app.daysWritten()
	points := today.PointCount()

	ui.Column(c).Grow(1).MinWidth(0).Children(func() {
		ui.Column(c).Padding(32, 32, 10).Gap(3).Children(func() {
			ui.Text(c, app.greeting(c.Now().Hour())+"。").Font(fontSerif).
				FontSize(27).Bold().SingleLine()
			ui.Text(c, model.DisplayDateFull(app.today)+" · "+model.Weekday(date)).
				FontSize(12.5).TextColor(t.TextMuted).SingleLine()
		})

		ui.Scroll(c).Grow(1).Padding(8, 32, 28).Gap(16).Children(func() {
			app.hero(c, quoteOf(app.today))

			ui.Row(c).Gap(14).Children(func() {
				app.statCard(c, iconSparkle, "连续记录", plural(streak, "天"))
				app.statCard(c, iconBook, "累计记录", plural(totalDays, "天"))
				app.statCard(c, iconCheckCircle, "今日要点", plural(points, "条"))
			})

			ui.Row(c).Gap(16).AlignItems(ui.Stretch).Children(func() {
				ui.Column(c).Grow(1.25).MinWidth(0).Children(func() {
					app.todayLearning(c, today).FillHeight()
				})
				ui.Column(c).Grow(1).MinWidth(0).Children(func() {
					app.dueCard(c).FillHeight()
				})
			})

			app.recentStrip(c)
		})
	})
}

// greeting 给问候语加上称呼。
func (app *App) greeting(hour int) string {
	if name := app.greetingName(); name != "" {
		return greetingByHour(hour) + "，" + name
	}
	return greetingByHour(hour)
}

// hero 是首页顶部的大卡片：必应每日一图；图没到或取不到时用低调的纸色底。
func (app *App) hero(c *ui.Context, quote string) {
	t := c.Theme()
	box := ui.Box(c).Height(212).Radius(18).Clip()

	if app.bg != nil {
		box.Children(func() {
			// 绝对定位 + 负 ZIndex：照片与压暗层垫在正文下面。
			ui.Image(c, app.bg).Absolute().Top(0).Left(0).Right(0).Bottom(0).
				Fit(ui.Cover).ZIndex(-3)
			ui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).ZIndex(-2).
				Gradient(ui.Transparent, ui.RGBA(0, 0, 0, 0.78), 178)
			heroContent(c, quote, "Bing 每日一图 · "+app.bgCopyright, true)
		})
		return
	}

	// 没有图片：用主题的纸色底，而不是一大块强调色。
	box.Background(t.Surface).Border(1, t.Border)
	box.Children(func() {
		credit := ""
		if app.bgLoading {
			credit = "正在加载今日图片…"
		}
		heroContent(c, quote, credit, false)
	})
}

// heroContent 是卡片上的文字：标签、寄语、图片署名。
func heroContent(c *ui.Context, quote, credit string, onImage bool) {
	t := c.Theme()
	labelColor, quoteColor, creditColor := t.TextMuted, t.Text, t.TextMuted
	if onImage {
		labelColor = ui.RGBA(255, 255, 255, 0.85)
		quoteColor = ui.RGBA(255, 255, 255, 0.98)
		creditColor = ui.RGBA(255, 255, 255, 0.72)
	}
	ui.Column(c).Fill().Justify(ui.End).Padding(24, 26).Gap(9).Children(func() {
		ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Icon(c, iconSparkle).FontSize(12).TextColor(labelColor)
			ui.Text(c, "今日寄语").FontSize(10.5).FontWeight(600).
				LetterSpacing(1.2).TextColor(labelColor)
		})
		ui.Text(c, quote).Font(fontSerif).FontSize(21).FontWeight(600).LineHeight(1.45).
			TextColor(quoteColor).MaxWidth(660)
		if credit != "" {
			ui.Text(c, credit).FontSize(10).TextColor(creditColor).SingleLine().
				MaxWidth(660).Tooltip(credit)
		}
	})
}

// statCard 是一张小统计卡。
func (app *App) statCard(c *ui.Context, icon *ui.SVG, label, value string) {
	t := c.Theme()
	ui.Column(c).Grow(1).Padding(16).Gap(6).Radius(14).
		Background(cardBG(t)).Border(1, t.Border).
		Shadow(0, 1, 2, 0, ui.RGBA(0, 0, 0, 0.04)).
		Children(func() {
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, icon).FontSize(13).TextColor(accentInk(t))
				ui.Text(c, label).FontSize(11.5).TextColor(t.TextMuted).SingleLine()
			})
			ui.Text(c, value).TextColor(t.Text).Font(fontSerif).FontSize(22).Bold().
				SingleLine().FontFeatures("lnum")
		})
}

// todayLearning 首页左栏：今天记下的学习条目。
func (app *App) todayLearning(c *ui.Context, today model.Day) ui.Element {
	t := c.Theme()
	const max = 3

	return sectionCard(c, "今天学了什么", "", func() {
		if len(today.Entries) == 0 {
			emptyState(c, iconBook, "今天还没有记录",
				"记录科目、章节，以及今天弄懂的要点。",
				"写今天的日记", func() { app.openEditor(app.today, "") })
			return
		}

		for i, e := range today.Entries {
			if i == max {
				break
			}
			ui.Column(c).Gap(4).Children(func() {
				ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
					dot(c, 6, accentInk(t))
					ui.Text(c, e.Subject).FontSize(13).FontWeight(600).SingleLine()
					if topic := topicLine(e); topic != "" {
						ui.Text(c, topic).FontSize(11.5).
							TextColor(t.TextMuted).SingleLine().Grow(1)
					}
				})
				// 有要点就露第一条，只有备注就露备注。
				summary := ""
				if len(e.Points) > 0 {
					summary = e.Points[0].Text
				} else if strings.TrimSpace(e.Note) != "" {
					summary = firstLine(e.Note)
				}
				if summary != "" {
					ui.Text(c, summary).FontSize(12.5).TextColor(t.TextMuted).
						SingleLine().Padding(0, 14)
				}
			})
		}

		if extra := len(today.Entries) - max; extra > 0 {
			ui.Textf(c, "另有 %d 条记录", extra).FontSize(11.5).TextColor(t.TextMuted)
		}

		ui.Link(c, "在日记里查看全部", "/diary").FontSize(12).TextColor(accentInk(t))
	})
}

// dueCard 首页右栏：今天与明天到期的待办（含已过期的）。
func (app *App) dueCard(c *ui.Context) ui.Element {
	t := c.Theme()
	due := model.DueSoon(app.todos, app.today)

	return sectionCard(c, "待办提醒", "", func() {
		if len(due) == 0 {
			ui.Text(c, "近两天没有到期的待办。").FontSize(12.5).TextColor(t.TextMuted)
			ui.Row(c).Children(func() {
				ui.Button(c, "查看待办").OnClick(func() { app.router.Push("/todo") })
			})
			return
		}

		for i, todo := range due {
			if i == 4 {
				break
			}
			app.dueRow(c, todo)
		}

		if extra := len(due) - 4; extra > 0 {
			ui.Textf(c, "另有 %d 条", extra).FontSize(11.5).TextColor(t.TextMuted)
		}

		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Textf(c, "%d 条待处理", len(due)).FontSize(11.5).TextColor(t.TextMuted)
			ui.Spacer(c)
			ui.Link(c, "全部待办", "/todo").FontSize(12).TextColor(accentInk(t))
		})
	})
}

// dueRow 是一条到期提醒：勾选框、内容、截止状态。
func (app *App) dueRow(c *ui.Context, todo model.Todo) {
	t := c.Theme()
	id := todo.ID
	idx := app.todoIndex(id)
	if idx < 0 {
		return
	}

	ui.Row(c).Gap(9).AlignItems(ui.Center).Children(func() {
		cb := ui.Checkbox(c.Key("home-todo-"+id), &app.todos[idx].Done, "").Label(todo.Text)
		cb.OnChange(func() {
			if i := app.todoIndex(id); i >= 0 {
				app.saveTodo(app.todos[i])
			}
		})
		ui.Text(c, todo.Text).FontSize(12.5).Grow(1).SingleLine()

		state := todo.DueStateOf(app.today)
		label := todo.DueLabel(app.today)
		color := t.TextMuted
		switch state {
		case model.DueOverdue:
			color = t.Danger
		case model.DueToday:
			color = accentInk(t)
		}
		ui.Text(c, label).FontSize(10.5).TextColor(color)
	})
}

// recentStrip 是首页底部的最近 7 天。
func (app *App) recentStrip(c *ui.Context) {
	ui.Column(c).Gap(10).Children(func() {
		ui.Text(c, "最近 7 天").Font(fontSerif).FontSize(14).FontWeight(600)
		ui.Row(c).Gap(8).Children(func() {
			for i := 6; i >= 0; i-- {
				date := model.MustDate(app.today).AddDate(0, 0, -i)
				app.dayCell(c, date, i == 0)
			}
		})
	})
}

// dayCell 是最近 7 天里的一格。
//
// 数字放在一个居中的盒子里：文本自身不参与居中，
// 靠容器的 Center() 才能让 12px 的字在 32px 的格子里真正居中。
func (app *App) dayCell(c *ui.Context, date time.Time, isToday bool) {
	t := c.Theme()
	key := date.Format(model.DateLayout)
	has := app.hasEntries(key)

	fill, fg := t.Surface, t.TextMuted
	if has {
		fill, fg = t.Accent, t.AccentText
	} else if isToday {
		fg = t.Text
	}

	cell := ui.Column(c).Key("day-" + key).Width(46).Gap(6).AlignItems(ui.Center).Children(func() {
		ui.Text(c, model.WeekdayShort(date)).FontSize(10.5).TextColor(t.TextMuted)

		box := ui.Box(c).Size(32, 32).Radius(9).Background(fill).Center()
		if isToday {
			box.Border(1.5, accentInk(t))
		}
		box.Children(func() {
			ui.Text(c, strconv.Itoa(date.Day())).FontSize(12).FontWeight(600).TextColor(fg)
		})
	})
	cell.Tooltip(key + " · " + daySummary(app, key))
}

// daySummary 用一句话概括某天，用于提示气泡。
func daySummary(app *App, date string) string {
	i := app.dayIndex(date)
	if i < 0 || !app.days[i].HasEntries() {
		return "没有日记"
	}
	return plural(len(app.days[i].Entries), "条学习记录")
}

// streak 计算连续记录的天数。
//
// 今天还没写不算断：从昨天往前数，昨天写过就仍然连续。
func (app *App) streak() int {
	day := model.MustDate(app.today)
	if day.IsZero() {
		return 0
	}
	if !app.hasEntries(app.today) {
		day = day.AddDate(0, 0, -1)
	}
	n := 0
	for ; n < 3650; n++ {
		if !app.hasEntries(day.Format(model.DateLayout)) {
			break
		}
		day = day.AddDate(0, 0, -1)
	}
	return n
}
