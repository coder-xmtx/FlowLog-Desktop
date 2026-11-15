// Package resources 把 resources/ 目录里需要打包进程序的图片嵌进二进制。
//
// mygo build 打包应用时会忽略 resources 目录里的 .go 文件，
// 所以这个文件只影响 go build / go test，不会进最终产物。
package resources

import (
	"embed"
	"log"

	"github.com/egoist/mygo/ui"
)

// Icon 是应用图标（512×512 以内、透明底的 PNG，会被缩到 96 显示）。
//
//go:embed icon.png
var Icon []byte

//go:embed fonts/SourceHanSerifSC-Regular.otf
var fontFS embed.FS

// SerifFamily 是注册时使用的字体族名，供 view 包引用。
const SerifFamily = "MySerif"

func init() {
	data, err := fontFS.ReadFile("fonts/SourceHanSerifSC-Regular.otf")
	if err != nil {
		log.Fatalf("读取嵌入字体失败: %v", err)
	}
	if err := ui.RegisterFont(data, SerifFamily); err != nil {
		log.Fatalf("注册字体失败: %v", err)
	}
}
