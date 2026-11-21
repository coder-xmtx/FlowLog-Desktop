//go:build windows

package app

import (
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/egoist/mygo"
)

// refreshShellIcons 在版本变化后请 Windows 资源管理器刷新图标缓存。
//
// 应用更新会替换 exe，桌面快捷方式指的还是同一个 exe，但 Windows 的图标
// 缓存可能还显示旧图标；SHCNE_ASSOCCHANGED 通知资源管理器重新读取。
// 只在版本变化时做一次，避免每次启动都让桌面闪一下。
func refreshShellIcons() {
	dir, err := mygo.App.Path(mygo.PathCache)
	if err != nil {
		return
	}
	marker := filepath.Join(dir, "last-version")
	if prev, err := os.ReadFile(marker); err == nil && strings.TrimSpace(string(prev)) == Version {
		return
	}
	if err := os.WriteFile(marker, []byte(Version), 0o644); err != nil {
		return
	}
	shChangeNotify := syscall.NewLazyDLL("shell32.dll").NewProc("SHChangeNotify")
	// SHCNE_ASSOCCHANGED | SHCNF_IDLIST
	_, _, _ = shChangeNotify.Call(0x08000000, 0x0, 0, 0)
}
