package app

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
	// Version 与项目根目录 mygo.json 的 version 保持一致。
	Version    = "0.5.1"
	Name       = "流记"
	NameEn     = "FlowLog"
	Tagline    = "把每天的学习，写成一条清晰的流。"
	Author     = "Mixsu"
	RepoURL    = "https://github.com/coder-xmtx/FlowLog-Desktop"
	License    = "MIT"
	UpdateHost = "GitHub Releases"
)

// CmdName 是平台上的快捷键修饰键写法，用于界面上的提示文字。
var CmdName = func() string {
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

// Icon 返回应用图标的位图，失败时为 nil（界面会退回色块）。
func Icon() *ui.Bitmap { return iconOnce() }

// MyGoVersion 从构建信息里读出实际链接进来的 MyGo 版本，不写死。
func MyGoVersion() string {
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

// MyGoVersionText 返回可展示的 MyGo 版本。
func MyGoVersionText() string {
	if v := MyGoVersion(); v != "" && v != "(devel)" {
		return v
	}
	return "本地构建"
}

// GoVersion 返回编译这个程序用的 Go 版本。
func GoVersion() string {
	return strings.TrimPrefix(runtime.Version(), "go")
}

// PlatformName 返回当前平台的中文名。
func PlatformName() string {
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
