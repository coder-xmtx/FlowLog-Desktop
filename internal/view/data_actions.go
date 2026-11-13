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

// reload 重新从数据库读取全部数据（恢复或导入之后）。
func (app *App) reload() {
	days, err := app.store.Days()
	if err != nil {
		app.toastError("读取数据失败", err.Error())
		return
	}
	app.days = days
	app.ensureToday()

	// 设置也可能被备份覆盖（外观、称呼）。
	if st, err := app.store.Settings(); err == nil {
		app.settings = st
	}

	// 列表状态复位：数据换了一批，旧的滚动位置与选中项已无意义。
	app.diaryList = ui.ListState{}
	app.todoList = ui.ListState{}
	app.filtered = nil
	app.todoRows = nil
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
				Message: fmt.Sprintf("备份包含 %d 天数据。恢复会用备份覆盖当前全部日记与设置，且无法撤销。",
					len(b.Days)),
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
		app.toast("恢复完成", plural(len(b.Days), "天")+"数据已载入")
	})
}

// ---------------------------------------------------------------- 旧版 JSON

// chooseJSONDir 选择一个旧版 data 目录，把里面的 *.json 导入数据库。
func (app *App) chooseJSONDir() {
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Parent:      app.win,
			Title:       "选择旧版数据目录（一天一个 JSON 文件）",
			Directory:   true,
			ButtonLabel: "导入",
		})
		if err != nil {
			app.post(func() { app.toastError("无法打开文件夹对话框", err.Error()) })
			return
		}
		if len(paths) == 0 {
			return
		}
		days, bad, err := store.ImportJSONDir(paths[0])
		if err != nil {
			app.post(func() { app.toastError("读取目录失败", err.Error()) })
			return
		}
		if len(days) == 0 {
			app.post(func() { app.toast("没有找到可导入的日记", "目录里需要形如 2026-11-13.json 的文件") })
			return
		}
		app.post(func() {
			msg := fmt.Sprintf("找到 %d 天数据，同日期会覆盖数据库里的记录。", len(days))
			if len(bad) > 0 {
				msg += " 另有 " + joinNames(bad) + " 无法读取，将被跳过。"
			}
			app.confirm = confirmState{
				Open:    true,
				Title:   "导入旧版数据？",
				Message: msg,
				Yes:     "导入",
				YesFunc: func() { go app.importDays(days, bad) },
			}
		})
	}()
}

// importDays 在后台导入并把结果提示给用户。
func (app *App) importDays(days []model.Day, bad []string) {
	n, err := app.store.ImportDays(days)
	if err != nil {
		app.post(func() { app.toastError("导入失败", err.Error()) })
		return
	}
	app.post(func() {
		app.reload()
		desc := plural(n, "天") + "数据已导入"
		if len(bad) > 0 {
			desc += "；跳过 " + joinNames(bad)
		}
		app.toast("导入完成", desc)
	})
}
