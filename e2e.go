//go:build e2e

package main

// 这个文件只在带 e2e 标签构建时生效，用来让外部工具（mygo-automation）
// 读取并操作运行中的窗口：截图、点按、输入。
//
// 发布版不要带这个标签：
//
//	go build -tags e2e -o app.exe .
//	MYGO_AUTOMATION=/tmp/flowlog.sock ./app.exe
import _ "github.com/egoist/mygo/plugins/automation"
