package view

import (
	"fmt"
	"strings"

	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// card 返回一张卡片：圆角、细描边、极淡的投影。
func card(c *ui.Context, body func()) ui.Element {
	t := c.Theme()
	return ui.Column(c).Padding(18).Gap(12).Radius(14).
		Background(cardBG(t)).Border(1, t.Border).
		Shadow(0, 1, 2, 0, ui.RGBA(0, 0, 0, 0.04)).
		Children(body)
}

// cardBG 返回卡片的底色。
//
// 深色主题下卡片比页面底色略亮一点，才有"抬起一层"的层次；
// 浅色主题下保持和页面同色，靠描边与投影区分。
func cardBG(t *ui.Theme) ui.Color {
	if t.Dark {
		return t.Background.Mix(ui.Hex("#ffffff"), 0.04)
	}
	return t.Background
}

// pageHeader 是页面顶部的标题区：左边标题与副标题，右边操作。
func pageHeader(c *ui.Context, title, subtitle string, actions func()) ui.Element {
	t := c.Theme()
	return ui.Row(c).AlignItems(ui.Center).Gap(16).Padding(28, 32, 14).Children(func() {
		ui.Column(c).Gap(2).Grow(1).MinWidth(0).Children(func() {
			// 标题用衬线体，和正文拉开层次。
			ui.Text(c, title).Font(fontSerif).FontSize(25).Bold().SingleLine()
			if subtitle != "" {
				ui.Text(c, subtitle).FontSize(12.5).TextColor(t.TextMuted).SingleLine()
			}
		})
		if actions != nil {
			actions()
		}
	})
}

// sectionCard 是一张带小标题的卡片。
func sectionCard(c *ui.Context, title, hint string, body func()) ui.Element {
	t := c.Theme()
	return card(c, func() {
		ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
			ui.Text(c, title).Font(fontSerif).FontSize(14).FontWeight(600)
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
		ui.Icon(c, icon).FontSize(28).TextColor(t.TextMuted).Opacity(0.7)
		ui.Text(c, title).Font(fontSerif).FontSize(15).FontWeight(600).TextColor(t.Text)
		if desc != "" {
			ui.Text(c, desc).FontSize(12.5).TextColor(t.TextMuted).
				TextAlign(ui.Center).MaxWidth(360)
		}
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

// dot 是条目前的圆点。
func dot(c *ui.Context, size float32, color ui.Color) ui.Element {
	return ui.Box(c).Size(size, size).Radius(size / 2).Background(color).Shrink(0)
}

// kvRow 是设置页里的一行：左边标签，右边内容。
func kvRow(c *ui.Context, label string, value func()) ui.Element {
	t := c.Theme()
	return ui.Row(c).AlignItems(ui.Center).Gap(16).Padding(7, 0).Children(func() {
		ui.Text(c, label).FontSize(12.5).TextColor(t.TextMuted).Width(96).Shrink(0)
		value()
	})
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

// formatSize 把字节数写成人类可读的大小。
func formatSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}

// topicLine 把章、节拼成一行面包屑文本。
func topicLine(e model.Entry) string {
	return strings.Join(e.TopicParts(), "  ›  ")
}

// firstLine 取多行文本的第一行，用于列表里的一行摘要。
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}
