package view

import (
	"runtime"
	"runtime/debug"
	"strings"
	"sync"

	"flowlog-desktop/resources"

	"github.com/egoist/mygo/ui"
)

// 关于这一份应用的自述。想换成自己的，改这里就够了。
const (
	// appVersion 与项目根目录 mygo.json 的 version 保持一致。
	appVersion = "0.4.0"
	appName    = "流记"
	appNameEn  = "FlowLog"
	appTagline = "把每天的学习，写成一条清晰的流。"
	appAuthor  = "Mixsu"
	appRepoURL = "https://github.com/coder-xmtx/FlowLog-Desktop"
	appLicense = "MIT"
)

// cmdName 是平台上的快捷键修饰键写法，用于界面上的提示文字。
var cmdName = func() string {
	if runtime.GOOS == "darwin" {
		return "⌘"
	}
	return "Ctrl"
}()

// iconOnce 让应用图标只在第一次用到时解码一次。
var iconOnce = sync.OnceValue(func() *ui.Bitmap {
	bmp, err := ui.DecodeBitmap(resources.Icon)
	if err != nil {
		return nil
	}
	return bmp
})

// iconBitmap 返回应用图标的位图，失败时为 nil（界面会退回色块）。
func iconBitmap() *ui.Bitmap { return iconOnce() }

// mygoVersion 从构建信息里读出实际链接进来的 MyGo 版本，不写死。
func mygoVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, dep := range info.Deps {
		if dep.Path == "github.com/egoist/mygo" {
			return dep.Version
		}
	}
	return ""
}

// mygoVersionText 返回可展示的 MyGo 版本。
func mygoVersionText() string {
	if v := mygoVersion(); v != "" && v != "(devel)" {
		return v
	}
	return "本地构建"
}

// goVersion 返回编译这个程序用的 Go 版本。
func goVersion() string {
	return strings.TrimPrefix(runtime.Version(), "go")
}

// platformName 返回当前平台的中文名。
func platformName() string {
	switch runtime.GOOS {
	case "windows":
		return "Windows"
	case "darwin":
		return "macOS"
	case "linux":
		return "Linux"
	}
	return runtime.GOOS
}
