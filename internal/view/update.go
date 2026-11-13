package view

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/egoist/mygo"
)

// updateState 是自动更新界面的状态。
type updateState struct {
	Checking   bool
	Installing bool
	Progress   float64
	Err        string
	// Version / Notes 是查到的新版本信息，空表示已是最新。
	Version string
	Notes   string
	// up 是待安装的更新，只在主线程读写。
	up *mygo.Update
}

// supported 报告这个构建能不能自更新（开发版、包管理器安装的都不行）。
func updateSupported() bool { return mygo.Updater.Enabled() }

// checkForUpdatesLater 在启动后稍等一会儿再静默检查一次。
func (app *App) checkForUpdatesLater() {
	if !updateSupported() {
		return
	}
	time.AfterFunc(8*time.Second, func() {
		app.checkUpdates(false)
	})
}

// checkUpdates 检查新版本。manual 为 true 时把结果也反馈到界面上。
func (app *App) checkUpdates(manual bool) {
	if app.update.Checking {
		return
	}
	app.update.Checking = true
	app.update.Err = ""
	app.redraw()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()

		up, err := mygo.Updater.Check(ctx)
		apply := func() {
			app.update.Checking = false
			switch {
			case err != nil && manual:
				app.update.Err = updateErrText(err)
			case up == nil && manual:
				app.toast("已是最新版本", "当前版本 v"+appVersion)
			case up != nil:
				app.update.Version = up.Version
				app.update.Notes = up.Notes
				app.update.up = up
				if manual {
					app.toast("发现新版本", "v"+up.Version)
				}
			}
		}
		if app.win != nil {
			app.win.Update(apply)
		} else {
			apply()
		}
	}()
}

// installUpdate 下载并安装待安装的更新，完成后重启应用。
func (app *App) installUpdate() {
	up := app.update.up
	if up == nil || app.update.Installing {
		return
	}
	app.update.Installing = true
	app.update.Progress = 0
	app.redraw()

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		last := -1
		err := up.Install(ctx, func(done, total int64) {
			if total <= 0 || app.win == nil {
				return
			}
			pct := int(float64(done) * 100 / float64(total))
			if pct == last {
				return
			}
			last = pct
			app.win.Update(func() { app.update.Progress = float64(pct) / 100 })
		})

		if app.win == nil {
			return
		}
		app.win.Update(func() {
			app.update.Installing = false
			if err != nil {
				app.update.Err = err.Error()
				app.toastError("更新失败", err.Error())
				return
			}
			// Relaunch 要在主线程上做。
			mygo.App.Relaunch()
		})
	}()
}

// updateErrText 把更新错误翻成一句人能看懂的话。
func updateErrText(err error) string {
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
