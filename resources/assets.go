// Package resources 把随程序发布的小文件嵌进二进制。
//
// 只有一个图标：应用窗口的图标、侧边栏标志、关于页面的图都来自它。
// 字体不在这里：界面统一使用系统字体，见 internal/view/theme.go 的字号阶梯。
package resources

import _ "embed"

// Icon 是应用图标（512×512 以内、透明底的 PNG，会被缩到 96 显示）。
//
//go:embed icon.png
var Icon []byte
