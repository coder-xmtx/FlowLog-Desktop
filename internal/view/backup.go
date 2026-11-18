// 自动备份在界面层的部分：启动时补一份、按钮手动备份、设置页的状态文字。
//
// 真正的文件读写都在 internal/store/autobackup.go：这边只负责挑时机、
// 把耗时操作放到后台 goroutine，再用 Window.Update 回到主线程更新界面。
package view

import (
	"fmt"
	"path/filepath"
	"time"

	"flowlog-desktop/internal/store"
)

// backupDir 返回备份目录：与数据库放在一起的 backups/ 子目录。
func (app *App) backupDir() string {
	return store.AutoBackupDir(app.store.Path())
}

// autoBackup 在启动时补一份当天的备份。
//
// 每天只在第一次启动时备份：如果 backups/ 里已经有今天的文件就直接跳过。
// 整个过程在后台跑，失败只弹一条错误提示，不打断启动。
func (app *App) autoBackup() {
	if !app.settings.AutoBackup {
		return
	}
	dir := app.backupDir()
	now := time.Now()
	if !store.BackupDue(dir, now) {
		return
	}
	path, err := app.store.AutoBackup(dir, app.settings.BackupKeep, now)
	if err != nil {
		app.post(func() { app.toastError("自动备份失败", err.Error()) })
		return
	}
	// 成功时不打扰用户：设置页能看到「上次备份」的时间。
	_ = path
}

// backupNow 立即备份一次，并把结果告诉用户。
//
// 从按钮触发，写文件放在 goroutine 里，完成后回到主线程弹提示。
func (app *App) backupNow() {
	go func() {
		path, err := app.store.AutoBackup(app.backupDir(), app.settings.BackupKeep, time.Now())
		if err != nil {
			app.post(func() { app.toastError("备份失败", err.Error()) })
			return
		}
		name := filepath.Base(path)
		app.post(func() { app.toast("已备份", name) })
	}()
}

// backupStatus 描述最近一份备份，给设置页显示。
//
// 返回两段文字：第一段是「上次备份 …」，第二段是备份目录的绝对路径。
// 从来没有备份过时，第一段提示会自动备份会在下次启动时进行。
func (app *App) backupStatus() (line, dir string) {
	dir = app.backupDir()
	info, ok := store.LatestBackup(dir)
	if !ok {
		return "还没有备份过", dir
	}
	return fmt.Sprintf("上次备份 %s · %s",
		info.ModTime.Format("1 月 2 日 15:04"), formatSize(info.Size)), dir
}
