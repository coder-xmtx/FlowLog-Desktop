package view

import (
	"runtime"
	"runtime/debug"
	"strings"
)

// 关于这一份应用的自述。想换成自己的，改这里就够了。
const (
	// appVersion 与项目根目录 mygo.json 的 version 保持一致。
	appVersion = "0.2.0"
	appName    = "流记"
	appNameEn  = "FlowLog"
	appTagline = "把每天的学习，写成一条清晰的流。"
	// appAuthor 是「关于」页里署的名字，换成你自己。
	appAuthor = "Mixsu"
	// appRepoURL 是仓库地址，换成你自己的。
	appRepoURL = "https://github.com/coder-xmtx/flowlog-desktop"
	appLicense = "MIT"
)

// cmdName 是平台上的快捷键修饰键写法，用于界面上的提示文字。
var cmdName = func() string {
	if runtime.GOOS == "darwin" {
		return "⌘"
	}
	return "Ctrl"
}()

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
	// 用 replace 指向本地 checkout 时，版本会显示成空或 v0.0.0。
	return ""
}

// goVersion 返回编译这个程序用的 Go 版本。
func goVersion() string {
	return strings.TrimPrefix(runtime.Version(), "go")
}
