// Package resources 把 resources/ 目录里需要打包进程序的图片嵌进二进制。
//
// mygo build 打包应用时会忽略 resources 目录里的 .go 文件，
// 所以这个文件只影响 go build / go test，不会进最终产物。
package resources

import _ "embed"

// Icon 是应用图标（512×512 以内、透明底的 PNG，会被缩到 96 显示）。
//
//go:embed icon.png
var Icon []byte
