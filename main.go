package main

import (
	"log"
	"os"
	"path/filepath"

	"flowlog-desktop/internal/store"
	"flowlog-desktop/internal/view"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func main() {
	mygo.App.WhenReady(func() {
		// 1. 存储层：确保目录存在，读出所有日记与外观设置。
		dir, err := dataDir()
		if err != nil {
			log.Fatal(err)
		}
		s, err := store.New(dir)
		if err != nil {
			log.Fatal(err)
		}

		// 2. 视图状态：读数据、建路由、算今天的问候。
		app := view.NewApp(s)

		win := mygo.NewWindow(mygo.WindowOptions{
			Title:     "流记 · FlowLog",
			Width:     1200,
			Height:    820,
			MinWidth:  880,
			MinHeight: 560,
			// 记住窗口位置与大小（由 MyGo 存在用户数据目录里）。
			StateKey: "main",
			// 原生渲染：没有 webview，没有前端构建。
			Content: ui.View(app.View),
		})
		app.SetWindow(win)
		mygo.App.SetMenu(appMenu(app))
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

// dataDir 决定数据放在哪。
//
//   - FLOWLOG_DATA 指定时用它（测试、或想把数据放到同步盘里）；
//   - 开发时用项目里的 ./data，也就是 go tool mygo dev 时看到的那个；
//   - 打包之后放到系统的用户数据目录，避免往 Program Files 里写。
//
// 这里用 PathAppData（%AppData% / Application Support）自己拼目录名，
// 而不是 PathUserData：后者要读 App.Name()，在某些启动时机它还可能是空的。
//
// 无论哪种，「关于」页都会显示实际生效的绝对路径。
func dataDir() (string, error) {
	if dir := os.Getenv("FLOWLOG_DATA"); dir != "" {
		return dir, nil
	}
	if mygo.IsDev() {
		return "./data", nil
	}
	base, err := mygo.App.Path(mygo.PathAppData)
	if err != nil {
		// 拿不到系统的用户目录时退回当前目录：至少还能用。
		return "./data", nil
	}
	return filepath.Join(base, "FlowLog", "data"), nil
}

// appMenu 是应用菜单：文件里放应用自己的动作，编辑/视图/窗口沿用系统角色，
// 这样输入框里的复制、粘贴、撤销和全屏都开箱可用。
func appMenu(app *view.App) *mygo.Menu {
	return mygo.NewMenu([]*mygo.MenuItem{
		{Role: mygo.RoleAppMenu},
		{Label: "文件", Submenu: []*mygo.MenuItem{
			{Label: "写今天的日记", Accelerator: "CmdOrCtrl+N", Click: func(*mygo.MenuItem, *mygo.Window) {
				app.NewEntry()
			}},
			mygo.Separator(),
			{Label: "首页", Accelerator: "CmdOrCtrl+1", Click: func(*mygo.MenuItem, *mygo.Window) {
				app.ShowPage("/")
			}},
			{Label: "日记", Accelerator: "CmdOrCtrl+2", Click: func(*mygo.MenuItem, *mygo.Window) {
				app.ShowPage("/diary")
			}},
			{Label: "待办", Accelerator: "CmdOrCtrl+3", Click: func(*mygo.MenuItem, *mygo.Window) {
				app.ShowPage("/todo")
			}},
			mygo.Separator(),
			{Role: mygo.RoleClose},
			{Role: mygo.RoleQuit},
		}},
		{Role: mygo.RoleEditMenu},
		{Role: mygo.RoleViewMenu},
		{Role: mygo.RoleWindowMenu},
		{Label: "帮助", Submenu: []*mygo.MenuItem{
			{Label: "关于流记", Click: func(*mygo.MenuItem, *mygo.Window) {
				app.ShowPage("/about")
			}},
		}},
	})
}
