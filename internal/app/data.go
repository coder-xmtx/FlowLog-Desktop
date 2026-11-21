package app

import (
	"flowlog-desktop/internal/lang"
	"strings"
	"time"

	"flowlog-desktop/internal/model"
	"flowlog-desktop/internal/store"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// 数据的导入导出与恢复：都在后台跑，结果通过 Post 回到主线程。

// Reload 重新从数据库读取全部数据（恢复备份之后）。
func (a *App) Reload() {
	days, err := a.Store.Days()
	if err != nil {
		a.ToastError("读取学记失败", err.Error())
		return
	}
	a.Days = days
	a.EnsureToday()

	todos, err := a.Store.Todos()
	if err != nil {
		a.ToastError("读取待办失败", err.Error())
		return
	}
	a.Todos = todos

	notes, err := a.Store.Notes()
	if err != nil {
		a.ToastError("读取随笔失败", err.Error())
		return
	}
	a.Notes = notes

	// 设置也可能被备份覆盖（外观、称呼、自动备份）。
	if st, err := a.Store.Settings(); err == nil {
		a.Settings = st
	}

	// 列表状态复位：数据换了一批，旧的滚动位置与选中项已无意义。
	a.StudyList = ui.ListState{}
	a.TodoList = ui.ListState{}
	a.NoteList = ui.ListState{}
	a.Filtered = nil
	a.Visible = nil
}

// ExportBackup 选择位置并导出一份 JSON 备份。
//
// 文件对话框会阻塞，所以整段跑在 goroutine 里。
func (a *App) ExportBackup() {
	go func() {
		path, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
			Parent:      a.Win,
			Title:       "导出备份",
			ButtonLabel: "导出",
			DefaultPath: "flowlog-backup-" + model.Today() + ".json",
			Filters:     []mygo.FileFilter{{Name: "JSON 备份", Extensions: []string{"json"}}},
		})
		if err != nil {
			a.Post(func() { a.ToastError("无法打开保存对话框", err.Error()) })
			return
		}
		if path == "" {
			return // 用户取消
		}
		if err := a.Store.ExportFile(path); err != nil {
			a.Post(func() { a.ToastError("导出失败", err.Error()) })
			return
		}
		a.Post(func() { a.Toast("备份已导出", path) })
	}()
}

// ChooseBackup 选择一份备份，确认后替换全部数据。
func (a *App) ChooseBackup() {
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{
			Parent:  a.Win,
			Title:   "选择备份文件",
			Filters: []mygo.FileFilter{{Name: "JSON 备份", Extensions: []string{"json"}}},
		})
		if err != nil {
			a.Post(func() { a.ToastError("无法打开文件对话框", err.Error()) })
			return
		}
		if len(paths) == 0 {
			return
		}
		b, err := store.ReadBackup(paths[0])
		if err != nil {
			a.Post(func() {
				a.ToastError("这个文件不是流记的备份", "请选择由「导出备份」生成的文件")
			})
			return
		}
		a.Post(func() {
			a.Confirm = ConfirmState{
				Open:  true,
				Title: "从备份恢复？",
				Message: lang.Format("备份包含 %d 天记录与 %d 条待办。恢复会覆盖当前全部数据，且无法撤销。",
					len(b.Days), len(b.Todos)),
				Yes:     "恢复",
				YesFunc: func() { go a.restoreBackup(b) },
			}
		})
	}()
}

// restoreBackup 在后台写库，完成后刷新界面。
//
// 覆盖之前先自动存一份「恢复前」的备份：恢复本身不可撤销，
// 万一选错了文件，还能从备份目录里找回原来的数据。
func (a *App) restoreBackup(b *store.Backup) {
	if _, err := a.Store.BackupBeforeRestore(a.BackupDir(), time.Now()); err != nil {
		a.Post(func() { a.ToastError("恢复前的备份失败", err.Error()) })
		return
	}
	if err := a.Store.Restore(b); err != nil {
		a.Post(func() { a.ToastError("恢复失败", err.Error()) })
		return
	}
	a.Post(func() {
		a.Reload()
		a.Toast(lang.T("恢复完成"), lang.Format("%d 天记录 · %d 条待办", len(b.Days), len(b.Todos)))
	})
}

// ExtractReleaseNotes 从 CHANGELOG 里取出某个版本的说明。
//
// 参数 changelog 是整份 CHANGELOG.md，version 形如 "0.5.1"；
// 找不到这一节时返回空字符串。
func ExtractReleaseNotes(changelog, version string) string {
	lines := strings.Split(changelog, "\n")
	start := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") && strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")) == version {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return ""
	}
	var out []string
	for _, line := range lines[start:] {
		if strings.HasPrefix(strings.TrimSpace(line), "## ") {
			break
		}
		out = append(out, line)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}
