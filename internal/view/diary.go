package view

import (
	"fmt"
	"strings"
	"time"

	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// diary 是日记页：一列按日期倒序的卡片，右上角写新的一条。
func (app *App) diary(c *ui.Context) {
	app.filtered = app.filterDays()

	written, points := 0, 0
	for i := range app.days {
		if !app.days[i].HasContent() {
			continue
		}
		written++
		points += app.days[i].PointCount()
	}
	subtitle := fmt.Sprintf("共 %d 天 · %d 条要点", written, points)
	if q := strings.TrimSpace(app.query); q != "" {
		subtitle = fmt.Sprintf("「%s」的搜索结果 · %d 天", q, len(app.filtered))
	}

	ui.Column(c).Grow(1).MinWidth(0).Children(func() {
		pageHeader(c, "日记", subtitle, func() {
			app.searchInput(c)

			add := ui.PrimaryButton(c.Key("diary-add"), "").Label("写日记").
				Tooltip("记录今天的学习（" + cmdName + " + N）")
			add.Children(func() {
				ui.Icon(c, iconPlus).FontSize(14)
				ui.Text(c, "写日记").SingleLine()
			})
			add.OnClick(func() { app.openEditor(app.today, "") })
		})

		if len(app.filtered) == 0 {
			if strings.TrimSpace(app.query) != "" {
				emptyState(c, iconSearch, "没有匹配的记录",
					"可以试试科目、章节，或要点里的关键词。",
					"清空搜索", func() { app.query = "" })
			} else {
				emptyState(c, iconBook, "还没有记录",
					"点击右上角「写日记」，记录科目、章节和今天弄懂的要点。",
					"写今天的日记", func() { app.openEditor(app.today, "") })
			}
			return
		}

		// 列表按日期给每一行一个稳定 Key：滚动位置与行内控件状态跟着
		// 「那一天」走，而不是跟着下标走。
		app.diaryList.Key = func(i int) any {
			if i < 0 || i >= len(app.filtered) {
				return i
			}
			return app.days[app.filtered[i]].Date
		}
		ui.List(c, &app.diaryList, len(app.filtered), func(i int) {
			app.dayCard(c, app.filtered[i])
		}).Grow(1).Padding(2, 32, 26).Gap(14)
	})
}

// searchInput 是标题栏右侧的搜索框，⌘F 把焦点送进来。
func (app *App) searchInput(c *ui.Context) {
	f := ui.SearchField(c.Key("diary-search"), &app.query).
		Placeholder("搜索科目、章节或要点").Label("搜索日记").Width(230)
	if app.focusSearch {
		f.Focus()
		app.focusSearch = false
	}
}

// filterDays 按搜索词算出要显示的下标。
//
// 今天总是显示（哪怕还没写），因为它是落笔的入口；搜索时不显示空的今天。
func (app *App) filterDays() []int {
	app.filtered = app.filtered[:0]
	q := strings.ToLower(strings.TrimSpace(app.query))
	for i := range app.days {
		d := app.days[i]
		if q == "" {
			app.filtered = append(app.filtered, i)
			continue
		}
		if d.Date == app.today && !d.HasContent() {
			continue
		}
		if dayMatches(d, q) {
			app.filtered = append(app.filtered, i)
		}
	}
	return app.filtered
}

// dayMatches 报告某一天是否命中搜索词。
func dayMatches(d model.Day, q string) bool {
	if strings.Contains(strings.ToLower(d.Date), q) ||
		strings.Contains(strings.ToLower(model.DisplayDate(d.Date)), q) {
		return true
	}
	for _, e := range d.Entries {
		if strings.Contains(strings.ToLower(e.Subject), q) ||
			strings.Contains(strings.ToLower(e.Chapter), q) ||
			strings.Contains(strings.ToLower(e.Section), q) ||
			strings.Contains(strings.ToLower(e.Note), q) {
			return true
		}
		for _, p := range e.Points {
			if strings.Contains(strings.ToLower(p.Text), q) {
				return true
			}
		}
	}
	for _, todo := range d.Todos {
		if strings.Contains(strings.ToLower(todo.Text), q) {
			return true
		}
	}
	return false
}

// dayCard 画一天：日期头与若干条学习记录。
func (app *App) dayCard(c *ui.Context, di int) {
	t := c.Theme()
	if di < 0 || di >= len(app.days) {
		return
	}
	day := app.days[di]
	date := day.Date
	rel := model.RelativeDay(date)
	points := day.PointCount()

	cardEl := ui.Column(c).Key("day-card-"+date).Padding(18).Gap(12).Radius(14).
		Background(cardBG(t)).Border(1, t.Border).
		Shadow(0, 1, 2, 0, ui.RGBA(0, 0, 0, 0.04))

	cardEl.Children(func() {
		ui.Row(c).Gap(9).AlignItems(ui.Center).Children(func() {
			ui.Text(c, model.DisplayDate(date)).Font(fontSerif).
				FontSize(16).Bold().SingleLine()
			ui.Text(c, model.Weekday(model.MustDate(date))).FontSize(12).TextColor(t.TextMuted)
			if rel != "" {
				pill(c, rel, t.Accent.Alpha(0.13), t.Accent)
			}
			if points > 0 {
				if done := day.DonePoints(); done > 0 {
					ui.Textf(c, "%d 条要点 · 已复习 %d", points, done).
						FontSize(11.5).TextColor(t.TextMuted)
				} else {
					ui.Textf(c, "%d 条要点", points).FontSize(11.5).TextColor(t.TextMuted)
				}
			}
			ui.Spacer(c)
			// 删除整天是危险操作，只在鼠标移到卡片上时出现。
			if cardEl.Hovered() && day.HasContent() {
				iconButton(c, "del-"+date, "删除这一天的记录", iconTrash, func() {
					app.askDeleteDay(date)
				})
			}
		})

		if len(day.Entries) == 0 {
			ui.Text(c, "这一天还没有记录。").
				FontSize(12.5).TextColor(t.TextMuted).Padding(2, 0)
			return
		}

		for j := range day.Entries {
			if j > 0 {
				ui.Box(c).Height(1).Background(t.Border).Margin(0, 2)
			}
			app.entryBlock(c, di, j)
		}
	})
}

// entryBlock 是一条学习记录：科目、章节路径、要点清单、备注。
func (app *App) entryBlock(c *ui.Context, di, ei int) {
	t := c.Theme()
	e := &app.days[di].Entries[ei]
	date := app.days[di].Date
	id := e.ID

	entry := ui.Column(c).Key("entry-" + id).Gap(7)
	hovered := entry.Hovered()

	entry.Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			pill(c, e.Subject, t.Accent.Alpha(0.13), t.Accent)

			if topic := topicLine(*e); topic != "" {
				ui.Text(c, topic).FontSize(12).TextColor(t.TextMuted).
					SingleLine().Grow(1)
			} else {
				ui.Spacer(c)
			}

			if e.Minutes > 0 {
				ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
					ui.Icon(c, iconClock).FontSize(11.5).TextColor(t.TextMuted)
					ui.Text(c, formatMinutes(e.Minutes)).FontSize(11).TextColor(t.TextMuted)
				})
			}
			if e.At != "" {
				ui.Text(c, e.At).FontSize(11).TextColor(t.TextMuted)
			}

			// 编辑按钮一直在，平时半透明，鼠标移到这一条上才变亮。
			edit := iconButton(c, "edit-entry-"+id, "编辑这条记录", iconPen, func() {
				app.openEditor(date, id)
			})
			if !hovered {
				edit.Opacity(0.5)
			}
		})

		if len(e.Points) > 0 {
			ui.Column(c).Gap(3).Padding(0, 2).Children(func() {
				for k := range e.Points {
					app.pointRow(c, di, ei, k)
				}
			})
		}

		if strings.TrimSpace(e.Note) != "" {
			ui.Row(c).Gap(10).Padding(0, 4).Children(func() {
				ui.Box(c).Width(2).Radius(2).Background(t.Border).Shrink(0)
				ui.Text(c, e.Note).FontSize(12.5).TextColor(t.TextMuted).
					LineHeight(1.55).Grow(1)
			})
		}
	})
}

// pointRow 是一条要点：勾上表示已经复习。
//
// 勾选直接写进 model（&Points[k].Done），OnChange 再负责落库。
func (app *App) pointRow(c *ui.Context, di, ei, pi int) {
	t := c.Theme()
	p := &app.days[di].Entries[ei].Points[pi]
	date := app.days[di].Date

	ui.Row(c).Gap(9).AlignItems(ui.Start).Children(func() {
		cb := ui.Checkbox(c.Key("pt-"+p.ID), &p.Done, "").Label(p.Text)
		cb.Tooltip("标记为已复习")
		cb.OnChange(func() { app.persist(date) })

		text := ui.Text(c, p.Text).FontSize(12.5).LineHeight(1.45).Grow(1)
		if p.Done {
			text.TextColor(t.TextMuted).Strikethrough()
		} else {
			text.TextColor(t.Text)
		}
	})
}

// ---------------------------------------------------------------- 写日记弹窗

// openEditor 打开编辑器：entryID 为空表示新增。
func (app *App) openEditor(date, entryID string) {
	ed := &app.ed
	ed.Open = true
	ed.Date = model.MustDate(date)
	if ed.Date.IsZero() {
		ed.Date, _ = model.ParseDate(app.today)
	}
	ed.EntryID = entryID
	ed.Err = ""
	ed.Subject, ed.Chapter, ed.Section, ed.Points, ed.Note, ed.Minutes = "", "", "", "", "", 0

	if entryID == "" {
		return
	}
	day := app.day(date)
	i := day.EntryByID(entryID)
	if i < 0 {
		return
	}
	e := day.Entries[i]
	ed.Subject = e.Subject
	ed.Chapter = e.Chapter
	ed.Section = e.Section
	ed.Note = e.Note
	ed.Minutes = float64(e.Minutes)
	lines := make([]string, 0, len(e.Points))
	for _, p := range e.Points {
		lines = append(lines, p.Text)
	}
	ed.Points = strings.Join(lines, "\n")
}

// diaryEditor 是「写日记」的模态框。
func (app *App) diaryEditor(c *ui.Context) {
	t := c.Theme()
	if !app.ed.Open {
		return
	}
	editing := app.ed.EntryID != ""
	title := "写今天的日记"
	if editing {
		title = "编辑这条记录"
	} else if app.ed.Date.Format(model.DateLayout) != app.today {
		title = "补记 " + model.DisplayDate(app.ed.Date.Format(model.DateLayout))
	}

	ui.Modal(c, &app.ed.Open, func() {
		// 表单高度跟着窗口走：窗口矮时表单自己滚动，按钮始终留在底部。
		_, windowHeight := c.Size()
		formHeight := windowHeight - 320
		if formHeight < 160 {
			formHeight = 160
		}

		ui.Column(c).Width(520).Children(func() {
			// 标题
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Icon(c, iconPen).FontSize(15).TextColor(t.Accent)
				ui.Text(c, title).Font(fontSerif).FontSize(17).Bold().SingleLine()
				ui.Spacer(c)
				ui.Text(c, cmdName+" + Enter 保存").FontSize(10.5).TextColor(t.TextMuted)
				iconButton(c, "close-editor", "关闭", iconClose, func() { app.ed.Open = false })
			})

			ui.Scroll(c).MaxHeight(formHeight).Children(func() {
				ui.Column(c).Gap(14).Children(func() {
					// 日期 + 科目
					ui.Row(c).Gap(12).AlignItems(ui.Start).Children(func() {
						if editing {
							ui.Column(c).Width(168).Gap(6).Children(func() {
								ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
									ui.Icon(c, iconCalendar).FontSize(11.5).TextColor(t.TextMuted)
									ui.Text(c, "日期").FontSize(11.5).TextColor(t.TextMuted)
								})
								ui.Text(c, model.DisplayDate(app.ed.Date.Format(model.DateLayout))).
									FontSize(12.5).Padding(7, 0)
							})
						} else {
							ui.Column(c).Width(168).Children(func() {
								ui.Field(c, "日期", func() {
									ui.DateInput(c.Key("editor-date"), &app.ed.Date).Label("日期")
								})
							})
						}
						ui.Column(c).Grow(1).MinWidth(0).Children(func() {
							ui.Field(c, "科目", func() {
								ui.Autocomplete(c.Key("editor-subject"), &app.ed.Subject, app.subjects()).
									Placeholder("计算机组成原理").Label("科目").AutoFocus()
							})
						})
					})

					// 章、节分成两个可选输入框
					ui.Row(c).Gap(12).AlignItems(ui.Start).Children(func() {
						ui.Column(c).Grow(1).MinWidth(0).Children(func() {
							ui.Field(c, "章", func() {
								ui.TextInput(c.Key("editor-chapter"), &app.ed.Chapter).
									Label("章").Placeholder("第 4 章 存储系统")
							})
						})
						ui.Column(c).Grow(1).MinWidth(0).Children(func() {
							ui.Field(c, "节", func() {
								ui.TextInput(c.Key("editor-section"), &app.ed.Section).
									Label("节").Placeholder("第 8 节 虚拟存储器")
							})
						})
					})
					ui.Text(c, "章、节都可以留空，只写科目也能保存。").
						FontSize(10.5).TextColor(t.TextMuted)

					ui.Field(c, "要点", func() {
						ui.TextArea(c.Key("editor-points"), &app.ed.Points).Label("要点").
							Lines(5, 10).
							Placeholder("一行一条，例如：\n页式虚拟存储器的地址映射\n页表的结构与访问流程")
					})
					ui.Text(c, "每行一条，会自动去掉开头的「- 」。").
						FontSize(10.5).TextColor(t.TextMuted)

					// 时长 + 备注
					ui.Row(c).Gap(12).AlignItems(ui.Start).Children(func() {
						ui.Column(c).Width(168).Children(func() {
							ui.Field(c, "学习时长（分钟）", func() {
								ui.NumberInput(c.Key("editor-minutes"), &app.ed.Minutes, 0, 600, 5).
									Label("学习时长")
							})
						})
						ui.Column(c).Grow(1).MinWidth(0).Children(func() {
							ui.Field(c, "备注", func() {
								ui.TextArea(c.Key("editor-note"), &app.ed.Note).Label("备注").
									Lines(2, 4).Placeholder("可选")
							})
						})
					})

					if app.ed.Err != "" {
						ui.Text(c, app.ed.Err).FontSize(12).TextColor(t.Danger)
					}
				})
			})

			// 分隔线与底部动作：线贴在表单下，按钮再往下留出间距。
			ui.Box(c).Height(1).Background(t.Border).Margin(6, -2, 0)
			ui.Row(c).Gap(8).AlignItems(ui.Center).Padding(16, 0, 2).Children(func() {
				if editing {
					ui.Button(c, "删除这条").FontSize(12).TextColor(t.Danger).OnClick(func() {
						date := app.ed.Date.Format(model.DateLayout)
						id := app.ed.EntryID
						app.ed.Open = false
						app.askDeleteEntry(date, id)
					})
				}
				ui.Spacer(c)
				ui.Button(c, "取消").OnClick(func() { app.ed.Open = false })
				ui.PrimaryButton(c, "保存").OnClick(app.saveEditor)
			})
		})

		if c.Shortcut(ui.Cmd, ui.KeyEnter) {
			app.saveEditor()
		}
	})
}

// saveEditor 校验草稿并写入某一天。
func (app *App) saveEditor() {
	subject := strings.TrimSpace(app.ed.Subject)
	if subject == "" {
		app.ed.Err = "请填写科目，例如「计算机组成原理」。"
		return
	}
	points := parsePointLines(app.ed.Points)
	note := strings.TrimSpace(app.ed.Note)
	if len(points) == 0 && note == "" {
		app.ed.Err = "至少填写一条要点，或写一段备注。"
		return
	}

	date := app.ed.Date.Format(model.DateLayout)
	di := app.dayIndex(date)
	if di < 0 {
		app.upsert(model.Day{Date: date, Entries: []model.Entry{}, Todos: []model.Todo{}})
		di = app.dayIndex(date)
		if di < 0 {
			app.ed.Err = "找不到 " + date + " 这一天，保存失败。"
			return
		}
	}

	entry := model.Entry{
		ID:      app.ed.EntryID,
		Subject: subject,
		Chapter: strings.TrimSpace(app.ed.Chapter),
		Section: strings.TrimSpace(app.ed.Section),
		Points:  points,
		Note:    note,
		Minutes: int(app.ed.Minutes),
		At:      time.Now().Format("15:04"),
	}

	if entry.ID == "" {
		entry.ID = model.NewID()
		app.days[di].Entries = append(app.days[di].Entries, entry)
		app.persist(date)
		app.ed.Open = false
		app.toast("已记录", subject+" · "+plural(len(points), "条要点"))
		return
	}

	j := app.days[di].EntryByID(entry.ID)
	if j < 0 {
		app.ed.Err = "这条记录已不存在，可能刚被删除。"
		return
	}
	old := app.days[di].Entries[j]
	entry.At = old.At
	// 按文本对上旧要点，带过「已复习」与 ID，避免编辑一次就丢勾。
	for i := range entry.Points {
		for _, p := range old.Points {
			if p.Text == entry.Points[i].Text {
				entry.Points[i].ID = p.ID
				entry.Points[i].Done = p.Done
				break
			}
		}
	}
	app.days[di].Entries[j] = entry
	app.persist(date)
	app.ed.Open = false
	app.toast("已更新", subject+" · "+plural(len(entry.Points), "条要点"))
}

// parsePointLines 把多行文本拆成要点，顺手去掉 Markdown 的列表符号。
func parsePointLines(s string) []model.Point {
	var out []model.Point
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		for _, prefix := range []string{"- ", "* ", "• ", "· "} {
			line = strings.TrimPrefix(line, prefix)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, model.Point{ID: model.NewID(), Text: line})
	}
	return out
}

// subjects 汇总出现过的科目，用于输入框补全。
func (app *App) subjects() []string {
	seen := make(map[string]bool, 8)
	out := make([]string, 0, 8)
	for _, d := range app.days {
		for _, e := range d.Entries {
			if e.Subject == "" || seen[e.Subject] {
				continue
			}
			seen[e.Subject] = true
			out = append(out, e.Subject)
		}
	}
	return out
}

// ---------------------------------------------------------------- 删除

// askDeleteEntry 弹出确认，删掉某一天里的某一条。
func (app *App) askDeleteEntry(date, entryID string) {
	day := app.day(date)
	i := day.EntryByID(entryID)
	if i < 0 {
		return
	}
	e := day.Entries[i]
	app.confirm = confirmState{
		Open:    true,
		Title:   "删除这条记录？",
		Message: e.Subject + " · " + plural(len(e.Points), "条要点") + "，删除后无法恢复。",
		Yes:     "删除",
		YesFunc: func() { app.deleteEntry(date, entryID) },
	}
}

// askDeleteDay 弹出确认，删掉一整天的记录。
func (app *App) askDeleteDay(date string) {
	day := app.day(date)
	what := plural(len(day.Entries), "条学习记录")
	if _, total := day.TodoProgress(); total > 0 {
		what += "、" + plural(total, "条待办")
	}
	app.confirm = confirmState{
		Open:    true,
		Title:   "删除 " + model.DisplayDate(date) + " 的全部记录？",
		Message: what + "，删除后无法恢复。",
		Yes:     "删除",
		YesFunc: func() { app.deleteDay(date) },
	}
}

// deleteEntry 删除一条记录，并给一次撤销的机会。
func (app *App) deleteEntry(date, entryID string) {
	di := app.dayIndex(date)
	if di < 0 {
		return
	}
	j := app.days[di].EntryByID(entryID)
	if j < 0 {
		return
	}
	removed := app.days[di].Entries[j]
	at := j
	app.days[di].Entries = append(app.days[di].Entries[:j], app.days[di].Entries[j+1:]...)
	app.persist(date)

	app.toastUndo("已删除「"+removed.Subject+"」", "撤销", func() {
		k := app.dayIndex(date)
		if k < 0 {
			app.upsert(model.Day{Date: date, Entries: []model.Entry{}, Todos: []model.Todo{}})
			k = app.dayIndex(date)
			if k < 0 {
				return
			}
		}
		ents := app.days[k].Entries
		if at > len(ents) {
			at = len(ents)
		}
		ents = append(ents, model.Entry{})
		copy(ents[at+1:], ents[at:])
		ents[at] = removed
		app.days[k].Entries = ents
		app.persist(date)
	})
}

// deleteDay 清空一整天，并给一次撤销的机会。
func (app *App) deleteDay(date string) {
	di := app.dayIndex(date)
	if di < 0 {
		return
	}
	backup := app.days[di]
	app.days[di].Entries = []model.Entry{}
	app.days[di].Todos = []model.Todo{}
	app.persist(date)

	app.toastUndo("已清空 "+model.DisplayDate(date), "撤销", func() {
		app.upsert(backup)
		app.persist(date)
	})
}

// confirmDialog 是全局的确认框。
func (app *App) confirmDialog(c *ui.Context) {
	if !app.confirm.Open {
		return
	}
	st := &app.confirm
	switch ui.AlertDialog(c, &st.Open, st.Title, st.Message, st.Yes, "取消") {
	case 0:
		fn := st.YesFunc
		st.YesFunc, st.Open = nil, false
		if fn != nil {
			fn()
		}
	}
}
