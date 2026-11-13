package main

import (
	"flowlog-desktop/internal/view"
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func main() {
	a := &view.App{}
	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:     "FlowLog Desktop",
			Width:     1200,
			Height:    800,
			MinWidth:  600,
			MinHeight: 400,
			// 记住窗口位置
			StateKey: "main",
			// 使用原生渲染
			Content: ui.View(a.View),
		})
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
