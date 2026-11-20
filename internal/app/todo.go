package app

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"flowlog-desktop/internal/model"
)

// 待办的动作：添加、勾选、删除、清除已完成与截止日期。
//
// 待办与学记是两套数据：这里的写入不会碰 Days。

// SaveTodo 写入或更新一条待办，并同步内存里的副本。
//
// 返回是否写入成功；失败时会弹一条错误提示。
func (a *App) SaveTodo(todo model.Todo) bool {
	if err := a.Store.SaveTodo(todo); err != nil {
		a.ToastError("保存待办失败", err.Error())
		return false
	}
	if i := a.TodoIndex(todo.ID); i >= 0 {
		a.Todos[i] = todo
	} else {
		a.Todos = append(a.Todos, todo)
	}
	a.SortTodos()
	return true
}

// DeleteTodoRow 从数据库与内存里删掉一条待办。
func (a *App) DeleteTodoRow(id string) bool {
	if err := a.Store.DeleteTodo(id); err != nil {
		a.ToastError("删除待办失败", err.Error())
		return false
	}
	for i := range a.Todos {
		if a.Todos[i].ID == id {
			a.Todos = append(a.Todos[:i], a.Todos[i+1:]...)
			return true
		}
	}
	return true
}

// TodoIndex 返回某条待办在 Todos 里的下标。
func (a *App) TodoIndex(id string) int {
	for i := range a.Todos {
		if a.Todos[i].ID == id {
			return i
		}
	}
	return -1
}

// SortTodos 保持与数据库一致的顺序：未完成在前，按截止日期与创建时间。
func (a *App) SortTodos() {
	sort.SliceStable(a.Todos, func(i, j int) bool {
		x, y := a.Todos[i], a.Todos[j]
		if x.Done != y.Done {
			return !x.Done
		}
		if x.Due != y.Due {
			if x.Due == "" {
				return false
			}
			if y.Due == "" {
				return true
			}
			return x.Due < y.Due
		}
		return x.Created > y.Created
	})
}

// TodoRows 按当前筛选算出要显示的待办下标。
//
// 结果同时写回 Visible，供列表的稳定 Key 使用。
func (a *App) TodoRows() []int {
	a.Visible = a.Visible[:0]
	for i := range a.Todos {
		switch a.TodoFilter {
		case 1:
			if a.Todos[i].Done {
				continue
			}
		case 2:
			if !a.Todos[i].Done {
				continue
			}
		}
		a.Visible = append(a.Visible, i)
	}
	return a.Visible
}

// SetDue 把新待办的截止日期设为今天之后 n 天，并关闭弹层。
func (a *App) SetDue(n int) {
	a.TodoDue = model.MustDate(model.DayAfter(a.Today, n))
	a.DuePicker = false
	a.FocusInput = true
}

// SetWeekend 把截止日期设为本周日（已经是周日就今天）。
func (a *App) SetWeekend() {
	now := model.MustDate(a.Today)
	offset := (7 - int(now.Weekday())) % 7
	a.TodoDue = now.AddDate(0, 0, offset)
	a.DuePicker = false
	a.FocusInput = true
}

// ClearDue 清掉新待办的截止日期。
func (a *App) ClearDue() {
	a.TodoDue = time.Time{}
	a.DuePicker = false
	a.FocusInput = true
}

// AddTodo 把输入框里的内容加成一条待办。
func (a *App) AddTodo() {
	text := strings.TrimSpace(a.TodoDraft)
	if text == "" {
		return
	}
	todo := model.Todo{
		ID:      model.NewID(),
		Text:    text,
		Created: a.Today,
		At:      time.Now().Format("15:04"),
	}
	if !a.TodoDue.IsZero() {
		todo.Due = a.TodoDue.Format(model.DateLayout)
	}
	if !a.SaveTodo(todo) {
		return
	}
	a.TodoDraft = ""
	a.TodoDue = time.Time{}
}

// DeleteTodo 删除一条待办，并给一次撤销的机会。
func (a *App) DeleteTodo(id string) {
	idx := a.TodoIndex(id)
	if idx < 0 {
		return
	}
	backup := a.Todos[idx]
	if !a.DeleteTodoRow(id) {
		return
	}
	a.ToastUndo("已删除「"+backup.Text+"」", "撤销", func() {
		a.SaveTodo(backup)
	})
}

// ClearDone 移除已完成的待办，并给一次撤销的机会。
func (a *App) ClearDone() {
	var backup []model.Todo
	kept := make([]model.Todo, 0, len(a.Todos))
	for _, todo := range a.Todos {
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
		if !a.DeleteTodoRow(todo.ID) {
			return
		}
	}
	a.Todos = kept
	a.ToastUndo(fmt.Sprintf("已清除 %d 条已完成", len(backup)), "撤销", func() {
		for _, todo := range backup {
			a.SaveTodo(todo)
		}
	})
}

// ToggleDuePicker 打开或关闭截止日期弹层，并把日历游标对到已选日期。
func (a *App) ToggleDuePicker() {
	a.DuePicker = !a.DuePicker
	if !a.DuePicker {
		return
	}
	if a.TodoDue.IsZero() {
		a.CalView = model.MustDate(a.Today)
	} else {
		a.CalView = a.TodoDue
	}
}
