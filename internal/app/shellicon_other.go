//go:build !windows

package app

// refreshShellIcons 只在 Windows 上有意义：其他平台的安装包会在更新时
// 替换整个 .app / 目录，图标跟着走。
func refreshShellIcons() {}
