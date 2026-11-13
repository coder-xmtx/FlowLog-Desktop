package view

import (
	"fmt"
	"strings"
	"time"

	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// todo 是待办页：只管今天，输入、打勾、删除、筛选。
func (app *App) todo(c *ui.Context) {
	t := c.Theme()
	today := app.day(app.today)
	done, total := today.TodoProgress()

	subtitle := "今天还没有待办"
	if total > 0 {
		subtitle = fmt.Sprintf("已完成 %d / %d", done, total)
		if done == total {
			subtitle = "今天的待办已完成"
		}
	}

	ui.Column(c).Grow(1).MinWidth(0).Children(func() {
		pageHeader(c, "待办", subtitle, func() {
			ui.Segmented(c.Key("todo-filter"), &app.todoFilter, "全部", "进行中", "已完成").
				Label("筛选")
		})

		// 输入框
		ui.Row(c).Gap(10).Padding(2, 32, 12).Children(func() {
			input := ui.TextInput(c.Key("todo-input"), &app.todoDraft).
				Placeholder("添加今天要做的事，回车即可添加").Label("新待办").Grow(1)
			if input.Submitted() {
				app.addTodo()
			}
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

		rows := app.todoRowsForToday()
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
		ti := app.dayIndex(app.today)
		ui.Column(c).Padding(0, 32, 24).Children(func() {
			app.todoList.Key = func(i int) any {
				if ti < 0 || i < 0 || i >= len(app.todoRows) {
					return i
				}
				return app.days[ti].Todos[app.todoRows[i]].ID
			}
			ui.List(c, &app.todoList, len(rows), func(i int) {
				if i < 0 || i >= len(app.todoRows) {
					return
				}
				app.todoRow(c, app.todoRows[i])
			}).MaxHeight(maxList).Padding(6, 4).Radius(12).
				Background(cardBG(t)).Border(1, t.Border).
				Shadow(0, 1, 2, 0, ui.RGBA(0, 0, 0, 0.04)).Clip().
				// 只保留竖向滚动条：行内容永远不该横向溢出，
				// 否则悬停时会出现一条多余的横向滚动条。
				ScrollbarAxes(false, true)

			ui.Text(c, model.DisplayDateFull(app.today)+" · "+model.Weekday(model.MustDate(app.today))).
				FontSize(11).TextColor(t.TextMuted).Padding(10, 4, 0)
		})
	})
}

// todoRowsForToday 按当前筛选算出今天的待办下标。
func (app *App) todoRowsForToday() []int {
	app.todoRows = app.todoRows[:0]
	ti := app.dayIndex(app.today)
	if ti < 0 {
		return app.todoRows
	}
	for j, todo := range app.days[ti].Todos {
		switch app.todoFilter {
		case 1:
			if todo.Done {
				continue
			}
		case 2:
			if !todo.Done {
				continue
			}
		}
		app.todoRows = append(app.todoRows, j)
	}
	return app.todoRows
}

// todoEmptyState 是三种「没有内容」时各自的引导。
func (app *App) todoEmptyState(c *ui.Context, total int) {
	switch {
	case total == 0:
		emptyState(c, iconInbox, "今天还没有待办",
			"在上面的输入框里写下要做的事，回车即可添加。", "", nil)
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
// 行高固定为 34：鼠标移上去时右侧的删除按钮出现，如果行高跟着内容变，
// 列表会因为内容变高而冒出滚动条，鼠标一滑就闪。
func (app *App) todoRow(c *ui.Context, index int) {
	t := c.Theme()
	ti := app.dayIndex(app.today)
	if ti < 0 || index < 0 || index >= len(app.days[ti].Todos) {
		return
	}
	todo := &app.days[ti].Todos[index]

	row := ui.Row(c).Key("todo-row-"+todo.ID).Height(36).Gap(10).
		AlignItems(ui.Center).Padding(0, 10).Radius(8)
	row.Children(func() {
		cb := ui.Checkbox(c.Key("todo-cb-"+todo.ID), &todo.Done, "").Label(todo.Text)
		cb.Tooltip("完成 / 取消完成")
		cb.OnChange(func() { app.persist(app.today) })

		text := ui.Text(c, todo.Text).FontSize(13).Grow(1).SingleLine()
		if todo.Done {
			text.TextColor(t.TextMuted).Strikethrough()
		}

		if todo.At != "" {
			ui.Text(c, todo.At).FontSize(11).TextColor(t.TextMuted)
		}
		if row.Hovered() {
			del := ui.Button(c.Key("todo-del-"+todo.ID), "").
				Label("删除这条待办").Tooltip("删除").
				Size(24, 24).Padding(0).Radius(6).MinWidth(0)
			del.Children(func() { ui.Icon(c, iconTrash).FontSize(13) })
			del.OnClick(func() { app.deleteTodo(todo.ID) })
		}
	})
	if row.Hovered() {
		row.Background(t.Surface)
	}
}

// todoCheckbox 是首页「今日待办」里那个可以直接勾的复选行。
func (app *App) todoCheckbox(c *ui.Context, id string) {
	t := c.Theme()
	ti := app.dayIndex(app.today)
	if ti < 0 {
		return
	}
	for j := range app.days[ti].Todos {
		todo := &app.days[ti].Todos[j]
		if todo.ID != id {
			continue
		}
		ui.Row(c).Gap(9).AlignItems(ui.Center).Padding(1, 0).Children(func() {
			cb := ui.Checkbox(c.Key("home-todo-"+todo.ID), &todo.Done, "").Label(todo.Text)
			cb.OnChange(func() { app.persist(app.today) })
			text := ui.Text(c, todo.Text).FontSize(12.5).Grow(1).SingleLine()
			if todo.Done {
				text.TextColor(t.TextMuted).Strikethrough()
			}
		})
		return
	}
}

// addTodo 把输入框里的内容加成一条待办。
func (app *App) addTodo() {
	text := strings.TrimSpace(app.todoDraft)
	if text == "" {
		return
	}
	ti := app.dayIndex(app.today)
	if ti < 0 {
		app.ensureToday()
		ti = app.dayIndex(app.today)
		if ti < 0 {
			return
		}
	}
	app.days[ti].Todos = append(app.days[ti].Todos, model.Todo{
		ID:   model.NewID(),
		Text: text,
		At:   time.Now().Format("15:04"),
	})
	app.todoDraft = ""
	app.persist(app.today)
}

// deleteTodo 删除一条待办，并给一次撤销的机会。
func (app *App) deleteTodo(id string) {
	ti := app.dayIndex(app.today)
	if ti < 0 {
		return
	}
	for j, todo := range app.days[ti].Todos {
		if todo.ID != id {
			continue
		}
		at := j
		app.days[ti].Todos = append(app.days[ti].Todos[:j], app.days[ti].Todos[j+1:]...)
		app.persist(app.today)
		app.toastUndo("已删除「"+todo.Text+"」", "撤销", func() {
			k := app.dayIndex(app.today)
			if k < 0 {
				return
			}
			todos := app.days[k].Todos
			if at > len(todos) {
				at = len(todos)
			}
			todos = append(todos, model.Todo{})
			copy(todos[at+1:], todos[at:])
			todos[at] = todo
			app.days[k].Todos = todos
			app.persist(app.today)
		})
		return
	}
}

// clearDone 移除已完成的待办，并给一次撤销的机会。
func (app *App) clearDone() {
	ti := app.dayIndex(app.today)
	if ti < 0 {
		return
	}
	backup := app.days[ti].Todos
	kept := make([]model.Todo, 0, len(backup))
	removed := 0
	for _, todo := range backup {
		if todo.Done {
			removed++
			continue
		}
		kept = append(kept, todo)
	}
	if removed == 0 {
		return
	}
	app.days[ti].Todos = kept
	app.persist(app.today)
	app.toastUndo(fmt.Sprintf("已清除 %d 条已完成", removed), "撤销", func() {
		k := app.dayIndex(app.today)
		if k < 0 {
			return
		}
		app.days[k].Todos = backup
		app.persist(app.today)
	})
}
