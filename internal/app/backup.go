// 自动备份在应用层的部分：启动时补一份、按钮手动备份、设置页的状态文字。
//
// 文件读写都在 internal/store/autobackup.go：这边只负责挑时机、
// 把耗时操作放到后台 goroutine，再用 Post 回到主线程更新界面。
package app

import (
	"fmt"
	"path/filepath"
	"time"

	"flowlog-desktop/internal/store"
)

// BackupDir 返回备份目录：与数据库放在一起的 backups/ 子目录。
func (a *App) BackupDir() string {
	return store.AutoBackupDir(a.Store.Path())
}

// AutoBackup 在启动时补一份当天的备份。
//
// 每天只在第一次启动时备份：如果 backups/ 里已经有今天的文件就直接跳过。
// 失败只弹一条错误提示，不打断启动；成功时不打扰用户。
func (a *App) AutoBackup() {
	if !a.Settings.AutoBackup {
		return
	}
	dir := a.BackupDir()
	now := time.Now()
	if !store.BackupDue(dir, now) {
		return
	}
	if _, err := a.Store.AutoBackup(dir, a.Settings.BackupKeep, now); err != nil {
		a.ToastError("自动备份失败", err.Error())
	}
}

// BackupNow 立即备份一次，并把结果告诉用户。
//
// 从按钮触发，写文件放在 goroutine 里，完成后回到主线程弹提示。
func (a *App) BackupNow() {
	go func() {
		path, err := a.Store.AutoBackup(a.BackupDir(), a.Settings.BackupKeep, time.Now())
		if err != nil {
			a.Post(func() { a.ToastError("备份失败", err.Error()) })
			return
		}
		name := filepath.Base(path)
		a.Post(func() { a.Toast("已备份", name) })
	}()
}

// BackupStatus 描述最近一份备份，给设置页显示。
//
// 返回两段文字：第一段是「上次备份 …」，第二段是备份目录的绝对路径。
func (a *App) BackupStatus() (line, dir string) {
	dir = a.BackupDir()
	info, ok := store.LatestBackup(dir)
	if !ok {
		return "还没有备份过", dir
	}
	return fmt.Sprintf("上次备份 %s · %s", info.ModTime.Format("1 月 2 日 15:04"), FormatSize(info.Size)), dir
}

// FormatSize 把字节数写成人类可读的大小。
func FormatSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
