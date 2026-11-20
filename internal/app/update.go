package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/egoist/mygo"
)

// 自动更新：检查、下载安装、错误文案。
//
// 更新包由 MyGo 的更新机制签名与校验；开发构建与包管理器安装的版本不参与。

// UpdateSupported 报告这个构建能不能自更新。
func UpdateSupported() bool { return mygo.Updater.Enabled() }

// CheckForUpdatesLater 在启动后稍等一会儿再静默检查一次。
func (a *App) CheckForUpdatesLater() {
	if !UpdateSupported() {
		return
	}
	time.AfterFunc(8*time.Second, func() {
		a.CheckUpdates(false)
	})
}

// CheckUpdates 检查新版本。
//
// 参数 manual 为 true 时把结果也反馈到界面上（发现新版本、已是最新、失败原因）。
func (a *App) CheckUpdates(manual bool) {
	if a.Update.Checking {
		return
	}
	a.Update.Checking = true
	a.Update.Err = ""
	a.Redraw()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()

		up, err := mygo.Updater.Check(ctx)
		apply := func() {
			a.Update.Checking = false
			switch {
			case err != nil && manual:
				a.Update.Err = UpdateErrText(err)
			case up == nil && manual:
				a.Toast("已是最新版本", "当前版本 v"+Version)
			case up != nil:
				a.Update.Version = up.Version
				a.Update.Notes = up.Notes
				a.Update.available = up
				if manual {
					a.Toast("发现新版本", "v"+up.Version)
				}
			}
		}
		a.Post(apply)
	}()
}

// InstallUpdate 下载并安装待安装的更新，完成后重启应用。
func (a *App) InstallUpdate() {
	up := a.Update.available
	if up == nil || a.Update.Installing {
		return
	}
	a.Update.Installing = true
	a.Update.Progress = 0
	a.Redraw()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		last := -1
		err := up.Install(ctx, func(done, total int64) {
			if total <= 0 {
				return
			}
			pct := int(float64(done) * 100 / float64(total))
			if pct == last {
				return
			}
			last = pct
			a.Post(func() { a.Update.Progress = float64(pct) / 100 })
		})

		a.Post(func() {
			a.Update.Installing = false
			if err != nil {
				a.Update.Err = err.Error()
				a.ToastError("更新失败", err.Error())
				return
			}
			// Relaunch 要在主线程上做。
			mygo.App.Relaunch()
		})
	}()
}

// UpdateErrText 把更新错误翻成一句人能看懂的话。
func UpdateErrText(err error) string {
	switch {
	case err == nil:
		return ""
	case errors.Is(err, mygo.ErrUpdatesDisabled):
		return "当前构建不支持自动更新"
	case errors.Is(err, context.DeadlineExceeded):
		return "连接更新服务器超时"
	}
	return fmt.Sprintf("检查更新失败：%v", err)
}
