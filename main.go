// Package main 是流记的入口：打开数据库、开窗口、装配发布说明。
package main

import (
	_ "embed"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"flowlog-desktop/internal/store"
	"flowlog-desktop/internal/view"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	// MyGo 内置控件（日历、数字输入等）的中文文案。
	_ "github.com/egoist/mygo/locales/zh-Hans"
)

// changelog 是仓库根目录的 CHANGELOG.md：设置页会展示当前版本的说明。
//
//go:embed CHANGELOG.md
var changelog string

func main() {
	// 界面只有中文：声明语言后，日期、数字等内置控件也使用中文格式。
	mygo.App.SetLanguages("zh-Hans")

	var db *store.Store

	mygo.App.WhenReady(func() {
		// 1. 打开数据库（必要时建表或升级），读出全部日记、待办与设置。
		dir, err := dataDir()
		if err != nil {
			log.Fatal(err)
		}
		s, err := store.Open(filepath.Join(dir, store.FileName))
		if err != nil {
			log.Fatal(err)
		}
		db = s

		app, err := view.NewApp(s)
		if err != nil {
			log.Fatal(err)
		}
		app.SetReleaseNotes(changelog)

		// 2. 开窗口。原生渲染：没有 webview，也没有前端构建。
		win := mygo.NewWindow(mygo.WindowOptions{
			Title:     "流记 · FlowLog",
			Width:     1200,
			Height:    820,
			MinWidth:  880,
			MinHeight: 560,
			// 记住窗口位置与大小（由 MyGo 存在用户数据目录里）。
			StateKey: "main",
			Content:  ui.View(app.View),
		})
		app.SetWindow(win)
		app.Start()

		setMenu()
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
	if db != nil {
		_ = db.Close()
	}
}

// setMenu 去掉窗口内的菜单栏。
//
// 应用自己的动作都在界面里（侧边栏与设置页），快捷键由 ui.Context.Shortcut
// 处理，所以 Windows / Linux 上不需要菜单。macOS 的系统菜单栏无法去掉，
// 保留「应用 / 编辑 / 窗口」三个标准菜单即可：输入框的复制粘贴、
// ⌘Q 退出都靠它们。
func setMenu() {
	if runtime.GOOS != "darwin" {
		mygo.App.SetMenu(nil)
		return
	}
	mygo.App.SetMenu(mygo.NewMenu([]*mygo.MenuItem{
		{Role: mygo.RoleAppMenu},
		{Role: mygo.RoleEditMenu},
		{Role: mygo.RoleWindowMenu},
	}))
}

// dataDir 决定数据库放在哪。
//
//   - FLOWLOG_DATA 指定时用它（测试，或把数据放到同步盘里）；
//   - 开发时用项目里的 ./data，也就是 go tool mygo dev 时看到的那个；
//   - 打包之后放到系统的用户数据目录，避免往 Program Files 里写。
//
// 用 PathAppData（%AppData% / Application Support）自己拼目录名，
// 而不是 PathUserData：后者要读 App.Name()，某些启动时机它还是空的。
func dataDir() (string, error) {
	if dir := os.Getenv("FLOWLOG_DATA"); dir != "" {
		return dir, nil
	}
	if mygo.IsDev() {
		return "./data", nil
	}
	base, err := mygo.App.Path(mygo.PathAppData)
	if err != nil {
		return "./data", nil
	}
	return filepath.Join(base, "FlowLog", "data"), nil
}
