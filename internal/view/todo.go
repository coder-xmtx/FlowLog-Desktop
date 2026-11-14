package view

import (
	"fmt"
	"strings"
	"time"

	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// todo 是待办页：全部待办的列表，可按状态筛选。
func (app *App) todo(c *ui.Context) {
	t := c.Theme()
	done, total := model.TodoProgress(app.todos)

	subtitle := "还没有待办"
	if total > 0 {
		subtitle = fmt.Sprintf("已完成 %d / %d", done, total)
	}

	ui.Column(c).Grow(1).MinWidth(0).Children(func() {
		pageHeader(c, "待办", subtitle, func() {
			ui.Segmented(c.Key("todo-filter"), &app.todoFilter, "全部", "进行中", "已完成").
				Label("筛选")
		})

		// 输入行：内容 + 可选的截止日期 + 添加
		ui.Row(c).Gap(10).Padding(2, 32, 12).Children(func() {
			input := ui.TextInput(c.Key("todo-input"), &app.todoDraft).
				Placeholder("添加待办，回车即可添加").Label("新待办").Grow(1)
			if input.Submitted() {
				app.addTodo()
			}
			if app.focusInput {
				// 选完截止日期后把光标送回输入框，可以直接接着回车。
				input.Focus()
				app.focusInput = false
			}
			app.duePickerButton(c)
			add := ui.PrimaryButton(c, "添加").OnClick(app.addTodo)
			add.Disabled(strings.TrimSpace(app.todoDraft) == "")
		})

		// 进度
		if total > 0 {
			ui.Row(c).Gap(12).AlignItems(ui.Center).Padding(0, 32, 8).Children(func() {
				ui.Progress(c, float64(done)/float64(total)).Grow(1).Height(5)
				ui.Textf(c, "%d / %d", done, total).FontSize(11.5).TextColor(t.TextMuted)
				if done > 0 {
					ui.Button(c, "清除已完成").FontSize(11.5).
						Tooltip("移除已完成的待办，可撤销").
						OnClick(app.clearDone)
				}
			})
		}

		rows := app.todoRows()
		if len(rows) == 0 {
			app.todoEmptyState(c, total)
			return
		}

		// 列表放在一张卡片里，卡片贴着内容高度；待办多了才滚动。
		_, windowHeight := c.Size()
		maxList := windowHeight - 270
		if maxList < 160 {
			maxList = 160
		}
		ui.Column(c).Padding(0, 32, 24).Children(func() {
			app.todoList.Key = func(i int) any {
				if i < 0 || i >= len(app.visible) {
					return i
				}
				return app.todos[app.visible[i]].ID
			}
			ui.List(c, &app.todoList, len(rows), func(i int) {
				if i < 0 || i >= len(app.visible) {
					return
				}
				app.todoRow(c, app.visible[i])
			}).MaxHeight(maxList).Padding(6, 4).Radius(12).
				Background(cardBG(t)).Border(1, t.Border).
				Shadow(0, 1, 2, 0, ui.RGBA(0, 0, 0, 0.04)).Clip().
				// 只保留竖向滚动条：行内容永远不该横向溢出。
				ScrollbarAxes(false, true)

			ui.Text(c, "按截止日期排序，没有截止日期的排在最后。").
				FontSize(11).TextColor(t.TextMuted).Padding(10, 4, 0)
		})
	})
}

// todoRows 按当前筛选算出要显示的待办下标。
func (app *App) todoRows() []int {
	app.visible = app.visible[:0]
	for i := range app.todos {
		switch app.todoFilter {
		case 1:
			if app.todos[i].Done {
				continue
			}
		case 2:
			if !app.todos[i].Done {
				continue
			}
		}
		app.visible = append(app.visible, i)
	}
	return app.visible
}

// todoEmptyState 是各种「没有内容」时的引导。
func (app *App) todoEmptyState(c *ui.Context, total int) {
	switch {
	case total == 0:
		emptyState(c, iconInbox, "还没有待办",
			"写下一件要做的事，可以顺便设定截止日期。", "", nil)
	case app.todoFilter == 1:
		emptyState(c, iconCheckCircle, "进行中的待办已清空",
			"剩余的待办会显示在这里。", "查看已完成", func() { app.todoFilter = 2 })
	case app.todoFilter == 2:
		emptyState(c, iconClock, "还没有已完成的待办",
			"完成一项之后，它会出现在这里。", "查看全部", func() { app.todoFilter = 0 })
	default:
		emptyState(c, iconInbox, "没有匹配的待办", "", "查看全部", func() { app.todoFilter = 0 })
	}
}

// todoRow 是一条待办。
//
// 删除按钮始终占位，只在鼠标移到这一行时才显示：
// 如果按需插入按钮，行高会变，列表会跟着跳动。
func (app *App) todoRow(c *ui.Context, index int) {
	t := c.Theme()
	if index < 0 || index >= len(app.todos) {
		return
	}
	todo := &app.todos[index]
	id := todo.ID

	row := ui.Row(c).Key("todo-row-"+id).Height(36).Gap(10).
		AlignItems(ui.Center).Padding(0, 10).Radius(8)
	hovered := row.Hovered()

	row.Children(func() {
		cb := ui.Checkbox(c.Key("todo-cb-"+id), &todo.Done, "").Label(todo.Text)
		cb.Tooltip("完成 / 取消完成")
		cb.OnChange(func() {
			if i := app.todoIndex(id); i >= 0 {
				app.saveTodo(app.todos[i])
			}
		})

		text := ui.Text(c, todo.Text).FontSize(13).Grow(1).SingleLine()
		if todo.Done {
			text.TextColor(t.TextMuted).Strikethrough()
		}

		if todo.Due != "" {
			app.dueChip(c, *todo)
		}

		del := ui.Button(c.Key("todo-del-"+id), "").
			Label("删除这条待办").Tooltip("删除").
			Size(26, 26).Padding(0).Radius(6).MinWidth(0)
		del.Children(func() { ui.Icon(c, iconTrash).FontSize(13) })
		del.OnClick(func() { app.deleteTodo(id) })
		if !hovered {
			// 占着位置，但既不画出来也不接收指针。
			del.Invisible()
		}
	})
	if hovered {
		row.Background(t.Surface)
	}
}

// dueChip 是待办行尾的截止日期标记。
func (app *App) dueChip(c *ui.Context, todo model.Todo) {
	t := c.Theme()
	label := todo.DueLabel(app.today)
	fg := t.TextMuted
	switch todo.DueStateOf(app.today) {
	case model.DueOverdue:
		fg = t.Danger
	case model.DueToday:
		fg = accentInk(t)
	}
	if todo.Done {
		fg = t.TextMuted
	}
	ui.Text(c, label).FontSize(10.5).TextColor(fg).Shrink(0)
}

// duePickerButton 是输入行里的「截止日期」按钮与它的弹层。
func (app *App) duePickerButton(c *ui.Context) {
	t := c.Theme()
	label := "截止日期"
	if !app.todoDue.IsZero() {
		label = "截止 " + app.todoDue.Format("1月2日")
	}

	b := ui.Button(c.Key("todo-due"), "").Label("选择截止日期").Tooltip("设置截止日期").
		FontSize(12).Shrink(0)
	b.Children(func() {
		ui.Icon(c, iconCalendar).FontSize(13)
		ui.Text(c, label).SingleLine()
	})
	b.OnClick(func() {
		app.duePicker = !app.duePicker
		if app.duePicker {
			// 打开时把日历游标对到已选的截止日期，没有就落在今天。
			if app.todoDue.IsZero() {
				app.calView = model.MustDate(app.today)
			} else {
				app.calView = app.todoDue
			}
		}
	})

	ui.Popover(c, b, &app.duePicker, func() {
		ui.Column(c).Width(300).Gap(10).Padding(4).Children(func() {
			ui.Text(c, "截止日期").FontSize(11.5).FontWeight(600).TextColor(t.TextMuted)

			ui.Row(c).Gap(8).Children(func() {
				ui.Button(c, "今天").FontSize(11.5).OnClick(func() { app.setDue(0) })
				ui.Button(c, "明天").FontSize(11.5).OnClick(func() { app.setDue(1) })
				ui.Button(c, "本周末").FontSize(11.5).OnClick(app.setWeekend)
				ui.Spacer(c)
				ui.Button(c, "清除").FontSize(11.5).OnClick(func() {
					app.todoDue = time.Time{}
					app.duePicker = false
				})
			})

			// 日历只驱动 calView；在它上面选一天才算选定截止日期。
			cal := ui.Calendar(c.Key("due-calendar"), &app.calView).Label("截止日期")
			if cal.Changed() {
				app.todoDue = app.calView
				app.duePicker = false
			}
		})
	})
}

// setDue 把截止日期设为今天之后 n 天。
func (app *App) setDue(n int) {
	app.todoDue = model.MustDate(model.DayAfter(app.today, n))
	app.duePicker = false
	app.focusInput = true
}

// setWeekend 把截止日期设为本周日（已经是周日就今天）。
func (app *App) setWeekend() {
	now := model.MustDate(app.today)
	offset := (7 - int(now.Weekday())) % 7
	app.todoDue = now.AddDate(0, 0, offset)
	app.duePicker = false
	app.focusInput = true
}

// addTodo 把输入框里的内容加成一条待办。
func (app *App) addTodo() {
	text := strings.TrimSpace(app.todoDraft)
	if text == "" {
		return
	}
	todo := model.Todo{
		ID:      model.NewID(),
		Text:    text,
		Created: app.today,
		At:      time.Now().Format("15:04"),
	}
	if !app.todoDue.IsZero() {
		todo.Due = app.todoDue.Format(model.DateLayout)
	}
	if !app.saveTodo(todo) {
		return
	}
	app.todoDraft = ""
	app.todoDue = time.Time{}
}

// deleteTodo 删除一条待办，并给一次撤销的机会。
func (app *App) deleteTodo(id string) {
	idx := app.todoIndex(id)
	if idx < 0 {
		return
	}
	backup := app.todos[idx]
	if !app.deleteTodoRow(id) {
		return
	}
	app.toastUndo("已删除「"+backup.Text+"」", "撤销", func() {
		app.saveTodo(backup)
	})
}

// clearDone 移除已完成的待办，并给一次撤销的机会。
func (app *App) clearDone() {
	var backup []model.Todo
	kept := make([]model.Todo, 0, len(app.todos))
	for _, todo := range app.todos {
		if todo.Done {
			backup = append(backup, todo)
			continue
		}
		kept = append(kept, todo)
	}
	if len(backup) == 0 {
		return
	}
	for _, todo := range backup {
		if !app.deleteTodoRow(todo.ID) {
			return
		}
	}
	app.todos = kept
	app.toastUndo(fmt.Sprintf("已清除 %d 条已完成", len(backup)), "撤销", func() {
		for _, todo := range backup {
			app.saveTodo(todo)
		}
	})
}
