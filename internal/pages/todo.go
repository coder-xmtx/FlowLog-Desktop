package pages

import (
	"strings"

	"flowlog-desktop/internal/app"
	design "flowlog-desktop/internal/design"
	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// Todo 是待办页：全部待办的列表，可按状态筛选，输入行可以顺便设截止日期。
func Todo(c *ui.Context, a *app.App) {
	p := design.PaletteOf(c)
	done, total := model.TodoProgress(a.Todos)

	subtitle := "还没有待办"
	if total > 0 {
		subtitle = design.Tf("已完成 %d / %d", done, total)
	}

	ui.Column(c).Grow(1).MinWidth(0).Children(func() {
		design.PageHeader(c, "待办", subtitle, func() {
			ui.Segmented(c.Key("todo-filter"), &a.TodoFilter, design.T("全部"), design.T("进行中"), design.T("已完成")).
				Label(design.T("筛选"))
		})

		// 输入行：内容 + 可选的截止日期 + 添加
		ui.Row(c).Gap(10).Padding(2, 32, 12).Children(func() {
			input := ui.TextInput(c.Key("todo-input"), &a.TodoDraft).
				Placeholder("添加待办，回车即可添加").Label(design.T("新待办")).
				Height(design.ControlHeight).Grow(1)
			if input.Submitted() {
				a.AddTodo()
			}
			if a.FocusInput {
				// 选完截止日期后把光标送回输入框，可以直接接着回车。
				input.Focus()
				a.FocusInput = false
			}
			duePicker(c, a)
			add := design.FilledButton(c, "todo-add", "添加", design.IconPlus, a.AddTodo)
			add.Disabled(strings.TrimSpace(a.TodoDraft) == "")
		})

		// 进度
		if total > 0 {
			ui.Row(c).Gap(12).AlignItems(ui.Center).Padding(0, 32, 10).Children(func() {
				ui.Progress(c, float64(done)/float64(total)).Grow(1).Height(6)
				design.Textf(c, "%d / %d", done, total).FontSize(design.FontCaption).
					TextColor(p.OnSurfaceVariant)
				if done > 0 {
					design.TextButton(c, "clear-done", "清除已完成", a.ClearDone)
				}
			})
		}

		rows := a.TodoRows()
		if len(rows) == 0 {
			todoEmptyState(c, a, total)
			return
		}

		// 列表放在一张卡片里，卡片贴着内容高度；待办多了才滚动。
		_, windowHeight := c.Size()
		maxList := windowHeight - 300
		if maxList < 160 {
			maxList = 160
		}
		ui.Column(c).Padding(0, 32, 24).Children(func() {
			a.TodoList.Key = func(i int) any {
				if i < 0 || i >= len(a.Visible) {
					return i
				}
				return a.Todos[a.Visible[i]].ID
			}
			ui.List(c, &a.TodoList, len(rows), func(i int) {
				if i < 0 || i >= len(a.Visible) {
					return
				}
				todoRow(c, a, a.Visible[i])
			}).MaxHeight(maxList).Padding(8).Radius(design.RadiusMD).
				Background(cardSurface(c)).
				Shadow(0, 1, 2, 0, ui.RGBA(23, 32, 74, 0.04)).
				Shadow(0, 6, 16, 0, ui.RGBA(23, 32, 74, 0.05)).
				Clip().
				// 只保留竖向滚动条：行内容永远不该横向溢出。
				ScrollbarAxes(false, true)

			design.Text(c, "按截止日期排序，没有截止日期的排在最后").
				FontSize(design.FontTiny).TextColor(p.OnSurfaceVariant).Padding(10, 4, 0)
		})
	})
}

// todoEmptyState 是各种「没有内容」时的引导。
func todoEmptyState(c *ui.Context, a *app.App, total int) {
	switch {
	case total == 0:
		design.EmptyState(c, design.IconInbox, "还没有待办",
			"写下一件要做的事，可以顺便设定截止日期", "", nil)
	case a.TodoFilter == 1:
		design.EmptyState(c, design.IconCheckCircle, "进行中的待办已清空",
			"剩余的待办会显示在这里", "查看已完成", func() { a.TodoFilter = 2 })
	case a.TodoFilter == 2:
		design.EmptyState(c, design.IconClock, "还没有已完成的待办",
			"完成一项之后，它会出现在这里", "查看全部", func() { a.TodoFilter = 0 })
	default:
		design.EmptyState(c, design.IconInbox, "没有匹配的待办", "", "查看全部", func() { a.TodoFilter = 0 })
	}
}

// todoRow 是一条待办。
//
// 删除按钮始终占位，只在鼠标移到这一行时才显示：
// 如果按需插入按钮，行高会变，列表会跟着跳动。
func todoRow(c *ui.Context, a *app.App, index int) {
	p := design.PaletteOf(c)
	if index < 0 || index >= len(a.Todos) {
		return
	}
	todo := &a.Todos[index]
	id := todo.ID

	row := ui.Row(c.Key("todo-row-"+id)).Height(44).Gap(12).
		AlignItems(ui.Center).Padding(0, 12).Radius(design.RadiusSM)
	hovered := row.Hovered()

	row.Children(func() {
		cb := design.Checkbox(c, "todo-cb-"+id, &todo.Done, todo.Text)
		cb.Tooltip(design.T("完成 / 取消完成"))
		cb.OnChange(func() {
			if i := a.TodoIndex(id); i >= 0 {
				a.SaveTodo(a.Todos[i])
			}
		})

		text := ui.Text(c, todo.Text).FontSize(design.FontBody).Grow(1).SingleLine().
			TextColor(p.OnSurface)
		if todo.Done {
			text.TextColor(p.OnSurfaceVariant).Strikethrough()
		}

		if todo.Due != "" {
			tone := design.ToneNeutral
			switch todo.DueStateOf(a.Today) {
			case model.DueOverdue:
				tone = design.ToneError
			case model.DueToday:
				tone = design.TonePrimary
			case model.DueTomorrow:
				tone = design.ToneSecondary
			}
			if todo.Done {
				tone = design.ToneNeutral
			}
			design.Chip(c, todo.DueLabel(a.Today), tone)
		}

		copyBtn := design.IconButton(c, "todo-copy-"+id, "复制这条待办", design.IconCopy, func() {
			if i := a.TodoIndex(id); i >= 0 {
				c.WriteClipboard(a.Todos[i].Text)
				a.Toast("已复制", a.Todos[i].Text)
			}
		})
		if !hovered {
			copyBtn.Invisible()
		}

		del := design.IconButton(c, "todo-del-"+id, "删除这条待办", design.IconTrash, func() {
			a.DeleteTodo(id)
		})
		if !hovered {
			// 占着位置，但既不画出来也不接收指针。
			del.Invisible()
		}
	})
	row.Background(design.HoverLayer(row, "hover", p.OnSurface, hovered))
}

// duePicker 是输入行里的「截止日期」按钮与它的弹层。
//
// 弹层左边是快捷选项，右边是日历：日历用 Browse 模式，翻月份只移动视图，
// 只有点具体某一天才会选定日期并关掉弹层——否则翻月会被当成选日期。
func duePicker(c *ui.Context, a *app.App) {
	p := design.PaletteOf(c)
	label := "截止日期"
	if !a.TodoDue.IsZero() {
		label = "截止 " + a.TodoDue.Format("1月2日")
	}

	anchor := design.OutlinedButton(c, "todo-due", label, design.IconCalendar, a.ToggleDuePicker)
	if !a.DuePicker {
		return
	}

	ui.Popover(c, anchor, &a.DuePicker, func() {
		ui.Row(c).Gap(18).Padding(10).Children(func() {
			// 左边：快捷选项
			ui.Column(c).Width(120).Gap(4).Children(func() {
				design.Text(c, "截止日期").FontSize(design.FontCaption).FontWeight(600).
					TextColor(p.OnSurfaceVariant).Padding(0, 10, 6)
				design.MenuButton(c, "due-today", "今天", func() { a.SetDue(0) })
				design.MenuButton(c, "due-tomorrow", "明天", func() { a.SetDue(1) })
				design.MenuButton(c, "due-weekend", "本周末", a.SetWeekend)
				design.Divider(c)
				design.MenuButton(c, "due-clear", "清除", a.ClearDue)
			})

			// 右边：日历只驱动 CalView；Browse 让翻月不选日期。
			cal := ui.Calendar(c.Key("due-calendar"), &a.CalView).
				Label("截止日期").Browse()
			if cal.Changed() {
				a.TodoDue = a.CalView
				a.DuePicker = false
				a.FocusInput = true
			}
		})
	})
}
