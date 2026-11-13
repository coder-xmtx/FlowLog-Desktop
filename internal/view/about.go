package view

import (
	"path/filepath"
	"runtime"
	"sync"

	"flowlog-desktop/resources"

	"github.com/egoist/mygo/ui"
)

// iconOnce 让图标只在第一次用到时解码一次。
var iconOnce = sync.OnceValues(func() (*ui.Bitmap, error) {
	return ui.DecodeBitmap(resources.Icon)
})

// about 是「关于」页：应用自述、开发者、数据位置、快捷键。
func (app *App) about(c *ui.Context) {
	t := c.Theme()

	ui.Column(c).Grow(1).MinWidth(0).Children(func() {
		ui.Scroll(c).Grow(1).Padding(34, 32, 30).Children(func() {
			ui.Column(c).MaxWidth(680).Gap(18).Children(func() {
				app.aboutHeader(c)
				app.aboutIntro(c)
				app.aboutDeveloper(c)
				app.aboutData(c)
				app.aboutKeys(c)

				ui.Text(c, "用 Go 和 MyGo 手写，没有一行 HTML。").
					FontSize(11).TextColor(t.TextMuted).Padding(4, 0)
			})
		})
	})
}

// aboutHeader 是应用图标、名字与版本。
func (app *App) aboutHeader(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).Gap(18).AlignItems(ui.Center).Children(func() {
		if bmp, err := iconOnce(); err == nil && bmp != nil {
			// 有真图标就用真图标，圆角裁一下，和系统里的观感一致。
			ui.Image(c, bmp).Size(76, 76).Radius(20).Fit(ui.Contain).Shrink(0)
		} else {
			logo := ui.Box(c).Size(76, 76).Radius(20).Background(t.Accent).Shrink(0)
			logo.Draw(func(p *ui.Painter, r ui.Rect) {
				drawLogoMark(p, r, t.AccentText, 3)
			})
		}

		ui.Column(c).Gap(4).Children(func() {
			ui.Row(c).Gap(9).AlignItems(ui.Baseline).Children(func() {
				ui.Text(c, appName).FontSize(26).Bold()
				ui.Text(c, appNameEn).FontSize(13).TextColor(t.TextMuted).LetterSpacing(0.4)
				pill(c, "v"+appVersion, t.Accent.Alpha(0.13), t.Accent)
			})
			ui.Text(c, appTagline).FontSize(12.5).TextColor(t.TextMuted)
		})
	})
}

// aboutIntro 是一段产品自述。
func (app *App) aboutIntro(c *ui.Context) {
	t := c.Theme()
	sectionCard(c, "这是什么", "", func() {
		ui.Text(c, "流记是一个用来记录每天学习情况的小工具。它把「今天学了什么」拆成"+
			"科目、章节和要点三层：一条记录属于一个科目，一个章节路径，下面挂着几条"+
			"今天真正弄懂的要点。比起一大段 Markdown，这样更容易回头检索、按科目复习，"+
			"也更容易看出自己这几天到底推进了什么。").
			FontSize(12.5).TextColor(t.TextMuted).LineHeight(1.65)
		ui.Text(c, "待办是给自己留的下一步：写一句、打个勾，今天就算有着落。").
			FontSize(12.5).TextColor(t.TextMuted).LineHeight(1.65)
	})
}

// aboutDeveloper 是开发者、仓库、许可证、技术栈。
func (app *App) aboutDeveloper(c *ui.Context) {
	t := c.Theme()
	sectionCard(c, "开发者", "", func() {
		kvRow(c, "作者", func() {
			ui.Text(c, appAuthor).FontSize(12.5)
		})
		kvRow(c, "仓库", func() {
			ui.Link(c, appRepoURL, appRepoURL).FontSize(12.5)
		})
		kvRow(c, "许可证", func() {
			ui.Text(c, appLicense).FontSize(12.5)
		})
		kvRow(c, "界面", func() {
			ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
				ui.Text(c, "MyGo "+mygoOrUnknown()+" · 原生 UI（非 WebView）").FontSize(12.5)
			})
		})
		kvRow(c, "编译", func() {
			ui.Text(c, "Go "+goVersion()+" · "+platformName()).FontSize(12.5).
				TextColor(t.TextMuted)
		})
	})
}

// aboutData 说明数据存在哪、什么格式，并给一个复制路径的按钮。
func (app *App) aboutData(c *ui.Context) {
	t := c.Theme()
	dir, err := filepath.Abs(app.store.Dir())
	if err != nil {
		dir = app.store.Dir()
	}

	sectionCard(c, "数据", "一天一个 JSON 文件", func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Grow(1).Padding(7, 10).Radius(7).Background(t.Surface).Children(func() {
				ui.Text(c, dir).FontSize(11.5).Font("monospace").Selectable().SingleLine().
					Tooltip(dir)
			})
			copyBtn := ui.Button(c, "").Label("复制路径").FontSize(11.5).Shrink(0)
			copyBtn.Children(func() {
				ui.Icon(c, iconCopy).FontSize(12)
				ui.Text(c, "复制路径").SingleLine()
			})
			copyBtn.OnClick(func() {
				c.WriteClipboard(dir)
				app.toast("已复制数据目录", dir)
			})
		})
		ui.Text(c, "每天一个 <日期>.json，文件名就是日期；同目录下的 settings.json 存的是"+
			"外观设置。写入时先写临时文件再改名，中途崩溃也不会留下写坏的日记。").
			FontSize(11.5).TextColor(t.TextMuted).LineHeight(1.6)
		ui.Text(c, "整个目录可以直接放进 Git 或网盘同步。").
			FontSize(11.5).TextColor(t.TextMuted).LineHeight(1.6)
	})
}

// aboutKeys 列出快捷键。
func (app *App) aboutKeys(c *ui.Context) {
	keys := []struct{ Key, What string }{
		{cmdName + "+N", "写一条今天的记录"},
		{cmdName + "+1 / 2 / 3 / 4", "切换到 首页 / 日记 / 待办 / 关于"},
		{cmdName + "+F", "在日记里搜索"},
		{cmdName + "+Enter", "保存正在写的日记"},
		{"Esc", "关闭弹窗"},
	}
	sectionCard(c, "快捷键", "", func() {
		ui.Column(c).Gap(2).Children(func() {
			for _, k := range keys {
				kvRow(c, k.Key, func() {
					ui.Text(c, k.What).FontSize(12.5).TextColor(c.Theme().TextMuted)
				})
			}
		})
	})
}

// mygoOrUnknown 返回 MyGo 版本，读不到时给一个诚实的说法。
func mygoOrUnknown() string {
	if v := mygoVersion(); v != "" && v != "(devel)" {
		return v
	}
	return "（本地构建）"
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
