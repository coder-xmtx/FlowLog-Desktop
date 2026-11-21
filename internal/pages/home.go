// Package pages 是流记的四个页面：首页、学记、待办、设置。
//
// 每个页面都是「状态 → 元素树」的函数：参数 c 是当前帧的上下文，
// a 是应用状态。页面只读状态、调动作，不直接碰数据库。
package pages

import (
	"strconv"
	"strings"

	"flowlog-desktop/internal/app"
	design "flowlog-desktop/internal/design"
	"flowlog-desktop/internal/lang"
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

// QuoteOf 返回某个日期的寄语。
func QuoteOf(date string) string {
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

// Home 是首页：问候、寄语大图、三张统计卡、今天的记录与待办提醒、最近 7 天。
func Home(c *ui.Context, a *app.App) {
	p := design.PaletteOf(c)
	date := model.MustDate(a.Today)
	today := a.Day(a.Today)

	ui.Column(c).Grow(1).MinWidth(0).Children(func() {
		ui.Column(c).Padding(32, 32, 10).Gap(2).Children(func() {
			design.Text(c, greeting(a, c.Now().Hour())).FontSize(design.FontDisplay).
				FontWeight(600).LetterSpacing(-0.6).SingleLine().TextColor(p.OnSurface)
			design.Text(c, model.DisplayDateFull(a.Today)+" · "+model.Weekday(date)).
				FontSize(design.FontBody).TextColor(p.OnSurfaceVariant).SingleLine()
		})

		ui.Scroll(c).Scrollbars(false).Grow(1).Padding(6, 32, 28).Gap(18).Children(func() {
			hero(c, a, QuoteOf(a.Today))

			ui.Row(c).Gap(14).Children(func() {
				statCard(c, design.IconSparkle, "连续记录", statValue(a.Streak(), "天"))
				statCard(c, design.IconBook, "累计记录", statValue(a.DaysWritten(), "天"))
				statCard(c, design.IconCheckCircle, "今日要点", statValue(today.PointCount(), "条"))
			})

			ui.Row(c).Gap(16).AlignItems(ui.Stretch).Children(func() {
				ui.Column(c).Grow(1.25).MinWidth(0).Children(func() {
					todayLearning(c, a, today).FillHeight()
				})
				ui.Column(c).Grow(1).MinWidth(0).Children(func() {
					dueCard(c, a).FillHeight()
				})
			})
		})
	})
}

// greeting 给问候语加上称呼。
func greeting(a *app.App, hour int) string {
	if name := a.GreetingName(); name != "" {
		return lang.T(greetingByHour(hour)) + lang.Separator() + name
	}
	return lang.T(greetingByHour(hour))
}

// hero 是首页顶部的大卡片：必应每日一图；图没到或取不到时用品牌渐变兜底。
func hero(c *ui.Context, a *app.App, quote string) {
	p := design.PaletteOf(c)
	box := ui.Box(c).Height(228).Radius(design.RadiusLG).Clip()

	if a.BG != nil {
		box.Children(func() {
			// 绝对定位 + 负 ZIndex：照片与压暗层垫在正文下面。
			ui.Image(c, a.BG).Absolute().Top(0).Left(0).Right(0).Bottom(0).
				Fit(ui.Cover).ZIndex(-3)
			ui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).ZIndex(-2).
				Gradient(ui.Transparent, ui.RGBA(4, 8, 30, 0.74), 178)
			heroContent(c, quote, lang.Format("Bing 每日一图 · %s", a.BGCopyright), true)
		})
		return
	}

	// 没有图片：低饱和的品牌色渐变，比一整块强调色安静得多。
	box.Background(p.SurfaceContainerHigh).
		Background(ui.Transparent).
		Gradient(p.PrimaryContainer, p.SurfaceContainerHigh, 155)
	box.Children(func() {
		credit := ""
		if a.BGLoading {
			credit = lang.T("正在加载今日图片…")
		}
		heroContent(c, quote, credit, false)
	})
}

// heroContent 是卡片上的文字：标签、寄语、图片署名。
//
// 参数 onImage 为真时用白色文字（压在照片上），否则用主题文字色。
func heroContent(c *ui.Context, quote, credit string, onImage bool) {
	p := design.PaletteOf(c)
	labelColor, quoteColor, creditColor := p.OnSurfaceVariant, p.OnSurface, p.OnSurfaceVariant
	if onImage {
		labelColor = ui.RGBA(255, 255, 255, 0.82)
		quoteColor = ui.RGBA(255, 255, 255, 0.98)
		creditColor = ui.RGBA(255, 255, 255, 0.70)
	}
	ui.Column(c).Fill().Justify(ui.End).Padding(26, 28).Gap(10).Children(func() {
		ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Icon(c, design.IconSparkle).FontSize(design.FontSmall).TextColor(labelColor)
			design.Text(c, "今日寄语").FontSize(design.FontCaption).FontWeight(600).
				LetterSpacing(1.2).TextColor(labelColor)
		})
		design.Text(c, quote).FontSize(design.FontHero).FontWeight(600).
			LineHeight(1.4).TextColor(quoteColor).MaxWidth(680)
		if credit != "" {
			ui.Text(c, credit).FontSize(design.FontTiny).TextColor(creditColor).
				SingleLine().MaxWidth(680).Tooltip(credit)
		}
	})
}

// statCard 是一张统计卡：圆角图标 + 标签 + 数字。
func statCard(c *ui.Context, icon *ui.SVG, label, value string) {
	p := design.PaletteOf(c)
	ui.Column(c).Grow(1).Padding(18).Gap(12).Radius(design.RadiusMD).
		Background(cardSurface(c)).
		Shadow(0, 1, 2, 0, ui.RGBA(23, 32, 74, 0.04)).
		Children(func() {
			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
				ui.Box(c).Size(34, 34).Radius(17).Background(p.PrimaryContainer).Center().
					Children(func() {
						ui.Icon(c, icon).FontSize(design.FontLead).TextColor(p.OnPrimaryContainer)
					})
				design.Text(c, label).FontSize(design.FontCaption).FontWeight(500).
					TextColor(p.OnSurfaceVariant).SingleLine()
			})
			design.Text(c, value).FontSize(design.FontHero).FontWeight(600).
				TextColor(p.OnSurface).SingleLine().FontFeatures("lnum")
		})
}

// todayLearning 首页左栏：今天记下的学习条目。
func todayLearning(c *ui.Context, a *app.App, today model.Day) ui.Element {
	p := design.PaletteOf(c)
	const max = 3

	return design.SectionCard(c, "今天学了什么", "", func() {
		if len(today.Entries) == 0 {
			design.EmptyState(c, design.IconBook, "今天还没有记录",
				"记录科目、章节，以及今天弄懂的要点",
				"写今天的学记", func() { a.OpenEditor(a.Today, "") })
			return
		}

		for i, e := range today.Entries {
			if i == max {
				break
			}
			ui.Column(c).Gap(5).Children(func() {
				ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
					design.Chip(c, e.Subject, design.TonePrimary)
					if topic := design.TopicLine(e); topic != "" {
						ui.Text(c, topic).FontSize(design.FontSmall).
							TextColor(p.OnSurfaceVariant).SingleLine().Grow(1)
					}
				})
				summary := ""
				if len(e.Points) > 0 {
					summary = e.Points[0].Text
				} else if strings.TrimSpace(e.Note) != "" {
					summary = design.FirstLine(e.Note)
				}
				if summary != "" {
					ui.Text(c, summary).FontSize(design.FontBody).
						TextColor(p.OnSurfaceVariant).SingleLine().Padding(0, 2)
				}
			})
		}

		if extra := len(today.Entries) - max; extra > 0 {
			design.Textf(c, "另有 %d 条记录", extra).FontSize(design.FontCaption).
				TextColor(p.OnSurfaceVariant)
		}

		ui.Row(c).Children(func() {
			design.TextButton(c, "view-all-study", "在学记里查看全部", func() {
				a.Router.Push("/study")
			})
		})
	})
}

// dueCard 首页右栏：今天与明天到期的待办（含已过期的）。
func dueCard(c *ui.Context, a *app.App) ui.Element {
	p := design.PaletteOf(c)
	due := model.DueSoon(a.Todos, a.Today)

	return design.SectionCard(c, "待办提醒", "", func() {
		if len(due) == 0 {
			design.Text(c, "近两天没有到期的待办").FontSize(design.FontBody).
				TextColor(p.OnSurfaceVariant)
			ui.Row(c).Children(func() {
				design.TonalButton(c, "goto-todo", "查看待办", nil, func() { a.Router.Push("/todo") })
			})
			return
		}

		for i, todo := range due {
			if i == 4 {
				break
			}
			dueRow(c, a, todo)
		}

		if extra := len(due) - 4; extra > 0 {
			design.Textf(c, "另有 %d 条", extra).FontSize(design.FontCaption).TextColor(p.OnSurfaceVariant)
		}

		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			design.Textf(c, "%d 条待处理", len(due)).FontSize(design.FontCaption).
				TextColor(p.OnSurfaceVariant)
			ui.Spacer(c)
			design.TextButton(c, "all-todos", "全部待办", func() { a.Router.Push("/todo") })
		})
	})
}

// dueRow 是一条到期提醒：勾选框、内容、截止状态。
func dueRow(c *ui.Context, a *app.App, todo model.Todo) {
	p := design.PaletteOf(c)
	id := todo.ID
	idx := a.TodoIndex(id)
	if idx < 0 {
		return
	}

	row := ui.Row(c.Key("home-todo-"+id)).Gap(10).AlignItems(ui.Center).
		Padding(4, 8).Radius(design.RadiusXS)
	hovered := row.Hovered()
	row.Children(func() {
		cb := design.Checkbox(c, "home-cb-"+id, &a.Todos[idx].Done, todo.Text)
		cb.OnChange(func() {
			if i := a.TodoIndex(id); i >= 0 {
				a.SaveTodo(a.Todos[i])
			}
		})
		ui.Text(c, todo.Text).FontSize(design.FontBody).Grow(1).SingleLine().
			TextColor(p.OnSurface)

		state := todo.DueStateOf(a.Today)
		tone := design.ToneNeutral
		switch state {
		case model.DueOverdue:
			tone = design.ToneError
		case model.DueToday:
			tone = design.TonePrimary
		}
		design.Chip(c, todo.DueLabel(a.Today), tone)
	})
	row.Background(design.HoverLayer(row, "hover", p.OnSurface, hovered))
}

// cardSurface 返回卡片的底色：浅色是纯白，深色是高一层容器。
func cardSurface(c *ui.Context) ui.Color {
	p := design.PaletteOf(c)
	if p.Dark {
		return p.SurfaceContainer
	}
	return p.SurfaceContainerLowest
}

// plural 返回「3 条」这样的数量词。
func plural(n int, unit string) string { return strconv.Itoa(n) + " " + unit }

// statValue 把「数字 + 量词」写成当前语言的写法：中文「3 天」，英文 "3 days"。
func statValue(n int, unit string) string {
	if lang.IsEnglish() {
		switch unit {
		case "天":
			if n == 1 {
				return "1 day"
			}
			return strconv.Itoa(n) + " days"
		case "条":
			if n == 1 {
				return "1 entry"
			}
			return strconv.Itoa(n) + " entries"
		}
	}
	return plural(n, unit)
}
