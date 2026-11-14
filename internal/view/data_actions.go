package view

import (
	"fmt"

	"flowlog-desktop/internal/model"
	"flowlog-desktop/internal/store"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// post 把 fn 放到主线程执行并重绘；没有窗口时直接执行。
//
// 后台 goroutine 改界面状态（包括弹提示）一律经过它。
func (app *App) post(fn func()) {
	if app.win != nil {
		app.win.Update(fn)
		return
	}
	fn()
}

// reload 重新从数据库读取全部数据（恢复备份之后）。
func (app *App) reload() {
	days, err := app.store.Days()
	if err != nil {
		app.toastError("读取日记失败", err.Error())
		return
	}
	app.days = days
	app.ensureToday()

	todos, err := app.store.Todos()
	if err != nil {
		app.toastError("读取待办失败", err.Error())
		return
	}
	app.todos = todos

	// 设置也可能被备份覆盖（外观、称呼）。
	if st, err := app.store.Settings(); err == nil {
		app.settings = st
	}

	// 列表状态复位：数据换了一批，旧的滚动位置与选中项已无意义。
	app.diaryList = ui.ListState{}
	app.todoList = ui.ListState{}
	app.filtered = nil
	app.visible = nil
}

// ---------------------------------------------------------------- 导出

// exportBackup 选择位置并导出一份 JSON 备份。
//
// 文件对话框会阻塞，所以整段跑在 goroutine 里。
func (app *App) exportBackup() {
	go func() {
		path, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
			Parent:      app.win,
			Title:       "导出备份",
			ButtonLabel: "导出",
			DefaultPath: "flowlog-backup-" + model.Today() + ".json",
			Filters:     []mygo.FileFilter{{Name: "JSON 备份", Extensions: []string{"json"}}},
		})
		if err != nil {
			app.post(func() { app.toastError("无法打开保存对话框", err.Error()) })
			return
		}
		if path == "" {
			return // 用户取消
		}
		if err := app.store.ExportFile(path); err != nil {
			app.post(func() { app.toastError("导出失败", err.Error()) })
			return
		}
		app.post(func() { app.toast("备份已导出", path) })
	}()
}

// ---------------------------------------------------------------- 恢复

// chooseBackup 选择一份备份，确认后替换全部数据。
func (app *App) chooseBackup() {
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Parent:  app.win,
			Title:   "选择备份文件",
			Filters: []mygo.FileFilter{{Name: "JSON 备份", Extensions: []string{"json"}}},
		})
		if err != nil {
			app.post(func() { app.toastError("无法打开文件对话框", err.Error()) })
			return
		}
		if len(paths) == 0 {
			return
		}
		b, err := store.ReadBackup(paths[0])
		if err != nil {
			app.post(func() {
				app.toastError("这个文件不是流记的备份", "请选择由「导出备份」生成的文件")
			})
			return
		}
		app.post(func() {
			app.confirm = confirmState{
				Open:  true,
				Title: "从备份恢复？",
				Message: fmt.Sprintf("备份包含 %d 天日记与 %d 条待办。恢复会覆盖当前全部数据，且无法撤销。",
					len(b.Days), len(b.Todos)),
				Yes:     "恢复",
				YesFunc: func() { go app.restoreBackup(b) },
			}
		})
	}()
}

// restoreBackup 在后台写库，完成后刷新界面。
func (app *App) restoreBackup(b *store.Backup) {
	if err := app.store.Restore(b); err != nil {
		app.post(func() { app.toastError("恢复失败", err.Error()) })
		return
	}
	app.post(func() {
		app.reload()
		app.toast("恢复完成", fmt.Sprintf("%d 天日记 · %d 条待办", len(b.Days), len(b.Todos)))
	})
}
