package view

import (
	"fmt"
	"math"
	"strings"

	"github.com/egoist/mygo/ui"
)

// card 返回一张卡片：圆角、细描边、几乎看不见的投影，像纸片上的一张便签。
func card(c *ui.Context, body func()) ui.Element {
	t := c.Theme()
	return ui.Column(c).Padding(18).Gap(12).Radius(14).
		Background(t.Background).Border(1, t.Border).
		Shadow(0, 1, 2, 0, ui.RGBA(0, 0, 0, 0.04)).
		Children(body)
}

// pageHeader 是每个页面顶部的标题区：左边标题与副标题，右边操作。
func pageHeader(c *ui.Context, title, subtitle string, actions func()) ui.Element {
	t := c.Theme()
	return ui.Row(c).AlignItems(ui.Center).Gap(16).Padding(26, 32, 14).Children(func() {
		ui.Column(c).Gap(3).Grow(1).MinWidth(0).Children(func() {
			ui.Text(c, title).FontSize(24).Bold().SingleLine()
			if subtitle != "" {
				ui.Text(c, subtitle).FontSize(12.5).TextColor(t.TextMuted).SingleLine()
			}
		})
		if actions != nil {
			actions()
		}
	})
}

// pageBody 是可滚动的正文区域，左右留出和标题一样的边距。
func pageBody(c *ui.Context, body func()) ui.Element {
	return ui.Scroll(c).Grow(1).Padding(4, 32, 28).Gap(16).Children(body)
}

// sectionCard 是一张带小标题的卡片。
func sectionCard(c *ui.Context, title, hint string, body func()) ui.Element {
	t := c.Theme()
	return card(c, func() {
		ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
			ui.Text(c, title).FontSize(13.5).Bold()
			if hint != "" {
				ui.Text(c, hint).FontSize(11.5).TextColor(t.TextMuted)
			}
		})
		body()
	})
}

// emptyState 是空列表时的引导：图标、说明和一颗按钮。
func emptyState(c *ui.Context, icon *ui.SVG, title, desc, action string, on func()) ui.Element {
	t := c.Theme()
	return ui.Column(c).FillWidth().Center().Padding(28, 16).Gap(10).Children(func() {
		ui.Icon(c, icon).FontSize(30).TextColor(t.TextMuted).Opacity(0.7)
		ui.Text(c, title).FontSize(14).Bold().TextColor(t.Text)
		ui.Text(c, desc).FontSize(12.5).TextColor(t.TextMuted).TextAlign(ui.Center).MaxWidth(360)
		if action != "" && on != nil {
			ui.Button(c, action).OnClick(on)
		}
	})
}

// iconButton 是一个只有图标的方形按钮，label 同时作为无障碍名称与提示。
func iconButton(c *ui.Context, key any, label string, icon *ui.SVG, on func()) ui.Element {
	b := ui.Button(c.Key(key), "").Label(label).Tooltip(label).
		Size(28, 28).Padding(0).Radius(7).MinWidth(0)
	b.Children(func() {
		ui.Icon(c, icon).FontSize(15)
	})
	return b.OnClick(on)
}

// pill 是一小块带底色的短文本，例如科目名、日期标记。
func pill(c *ui.Context, text string, bg, fg ui.Color) ui.Element {
	return ui.Row(c).Padding(2.5, 8).Radius(6).Background(bg).
		AlignItems(ui.Center).Shrink(0).Children(func() {
		ui.Text(c, text).FontSize(11.5).FontWeight(600).TextColor(fg).SingleLine()
	})
}

// dot 是一个用于条目前的小圆点。
func dot(c *ui.Context, size float32, color ui.Color) ui.Element {
	return ui.Box(c).Size(size, size).Radius(size / 2).Background(color).Shrink(0)
}

// muted 是一行次要说明文字。
func muted(c *ui.Context, text string) ui.Element {
	return ui.Text(c, text).FontSize(12).TextColor(c.Theme().TextMuted).SingleLine()
}

// kvRow 是「关于」页里的一行：左边标签，右边内容。
func kvRow(c *ui.Context, label string, value func()) ui.Element {
	t := c.Theme()
	return ui.Row(c).AlignItems(ui.Center).Gap(16).Padding(6, 0).Children(func() {
		ui.Text(c, label).FontSize(12.5).TextColor(t.TextMuted).Width(76).Shrink(0)
		value()
	})
}

// drawWaves 在矩形里画三道正弦波，呼应用户的图标。
//
// 它只负责画：Draw 回调可能一帧被调用多次，不能在这里改应用状态。
func drawWaves(p *ui.Painter, r ui.Rect, color ui.Color, amp, thickness float32) {
	const steps = 72
	for i := 0; i < 3; i++ {
		base := r.Y + r.H*(0.30+0.19*float32(i))
		var wave ui.Path
		for k := 0; k <= steps; k++ {
			f := float32(k) / steps
			x := r.X - r.W*0.05 + r.W*1.1*f
			y := base + amp*float32(math.Sin(float64(f)*math.Pi*2.4+float64(i)*0.9))
			if k == 0 {
				wave.MoveTo(x, y)
			} else {
				wave.LineTo(x, y)
			}
		}
		p.StrokePath(&wave, thickness, color)
	}
}

// drawLogoMark 在方框里画小小的波浪标志，用于侧边栏与「关于」页。
func drawLogoMark(p *ui.Painter, r ui.Rect, color ui.Color, thickness float32) {
	const steps = 40
	for i := 0; i < 3; i++ {
		base := r.Y + r.H*(0.28+0.22*float32(i))
		var wave ui.Path
		for k := 0; k <= steps; k++ {
			f := float32(k) / steps
			x := r.X + r.W*f
			y := base + r.H*0.055*float32(math.Sin(float64(f)*math.Pi*2+float64(i)*0.55))
			if k == 0 {
				wave.MoveTo(x, y)
			} else {
				wave.LineTo(x, y)
			}
		}
		p.StrokePath(&wave, thickness, color)
	}
}

// formatMinutes 把分钟数写成「45 分钟」「1 小时 20 分」。
func formatMinutes(m int) string {
	switch {
	case m <= 0:
		return ""
	case m < 60:
		return fmt.Sprintf("%d 分钟", m)
	case m%60 == 0:
		return fmt.Sprintf("%d 小时", m/60)
	default:
		return fmt.Sprintf("%d 小时 %d 分", m/60, m%60)
	}
}

// plural 返回「3 条」这样的数量词。
func plural(n int, unit string) string {
	return fmt.Sprintf("%d %s", n, unit)
}

// firstLine 取多行文本的第一行，用于列表里的一行摘要。
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}
