// 这个文件放界面里反复用的小零件：卡片、页头、空状态、勾选框、圆点等。
//
// 零件只负责「长什么样」与基本的可访问性（Label / Tooltip），
// 不含业务逻辑；需要读写 App 状态的东西放在各页面文件里。
package view

import (
	"fmt"
	"strings"

	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// 勾选框的尺寸常量，用来把它和旁边的文字对齐。
const (
	// checkboxSize 是 MyGo 勾选框方框的边长（主题 Space(4)）。
	checkboxSize = 16
	// pointCheckboxOffset 是要点行里勾选框的下移量：让方框居中于第一行文字。
	//
	// 要点文字用 fsBody 字号与 1.45 的行高，第一行高 fsBody*1.45；
	// 方框比它矮，要在第一行里居中就下移一半的差值。
	pointCheckboxOffset = (fsBody*1.45 - checkboxSize) / 2

	// checkboxNudge 是勾选框的光学微调。
	//
	// 几何上让 16 DIP 的方框居中于文字行盒就够了，但中文的墨迹中心
	// 比行盒中心低约 1 DIP，方框按行盒居中时会显得偏高，往下挪 1 DIP
	// 才和文字「看起来」在一条线上。
	checkboxNudge = 1
)

// card 返回一张卡片：圆角、细描边、极淡的投影。
//
// 参数 body 在卡片内部构建子元素；卡片尺寸由内容决定。
func card(c *ui.Context, body func()) ui.Element {
	t := c.Theme()
	return ui.Column(c).Padding(18).Gap(12).Radius(14).
		Background(cardBG(t)).Border(1, t.Border).
		Shadow(0, 1, 2, 0, ui.RGBA(0, 0, 0, 0.04)).
		Children(body)
}

// cardBG 返回卡片的底色。
//
// 浅色主题下卡片比页面（暖纸色）更接近白色，叠出层次；
// 深色主题下卡片比页面略亮一点。参数 t 是当前主题。
func cardBG(t *ui.Theme) ui.Color {
	if t.Dark {
		return t.Background.Mix(ui.Hex("#ffffff"), 0.04)
	}
	return ui.Hex("#fffcf7")
}

// pageHeader 是页面顶部的标题区：左边标题与副标题，右边操作。
//
// 参数 title 是页面名，subtitle 是标题下的一行说明（可为空），
// actions 在右侧构建按钮与筛选控件（可为 nil）。
func pageHeader(c *ui.Context, title, subtitle string, actions func()) ui.Element {
	t := c.Theme()
	return ui.Row(c).AlignItems(ui.Center).Gap(16).Padding(28, 32, 14).Children(func() {
		ui.Column(c).Gap(2).Grow(1).MinWidth(0).Children(func() {
			ui.Text(c, title).FontSize(fsTitle).Bold().SingleLine()
			if subtitle != "" {
				ui.Text(c, subtitle).FontSize(fsBody).TextColor(t.TextMuted).SingleLine()
			}
		})
		if actions != nil {
			actions()
		}
	})
}

// sectionCard 是一张带小标题的卡片。
//
// 参数 title 是卡片标题，hint 是标题右侧的浅色补充（可为空），
// body 在标题下面构建内容。
func sectionCard(c *ui.Context, title, hint string, body func()) ui.Element {
	t := c.Theme()
	return card(c, func() {
		ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
			ui.Text(c, title).FontSize(fsLead).FontWeight(600)
			if hint != "" {
				ui.Text(c, hint).FontSize(fsCaption).TextColor(t.TextMuted)
			}
		})
		body()
	})
}

// emptyState 是空列表时的引导：图标、说明和一颗按钮。
//
// 参数 icon 是顶部图标，title 与 desc 是两行说明，
// action 是按钮文字（空字符串表示不显示按钮），on 是按钮动作。
func emptyState(c *ui.Context, icon *ui.SVG, title, desc, action string, on func()) ui.Element {
	t := c.Theme()
	return ui.Column(c).FillWidth().Center().Padding(8, 8).Gap(10).Children(func() {
		ui.Icon(c, icon).FontSize(28).TextColor(t.TextMuted).Opacity(0.7)
		ui.Text(c, title).FontSize(fsLead).FontWeight(600).TextColor(t.Text)
		if desc != "" {
			ui.Text(c, desc).FontSize(fsBody).TextColor(t.TextMuted).
				TextAlign(ui.Center).MaxWidth(360)
		}
		if action != "" && on != nil {
			ui.Button(c, action).OnClick(on)
		}
	})
}

// checkbox 构建一个和右侧文字视觉对齐的勾选框。
//
// 参数 key 是它在当前列表里的稳定身份（同一列表内唯一），checked 是绑定
// 的状态，label 是给读屏与自动化用、界面上不显示的文字。
//
// MyGo 的勾选框是一个 16×16 的方框，自己在内部按行盒居中；这里补一个
// 固定的光学偏移（见 checkboxNudge），让它和中文文字看起来在一条线上。
func checkbox(c *ui.Context, key any, checked *bool, label string) ui.Element {
	return ui.Checkbox(c.Key(key), checked, "").Label(label).
		Translate(0, checkboxNudge)
}

// iconButton 是一个只有图标的方形按钮，label 同时作为无障碍名称与提示。
//
// 参数 key 是稳定身份，icon 是图标，on 是点击动作。
func iconButton(c *ui.Context, key any, label string, icon *ui.SVG, on func()) ui.Element {
	b := ui.Button(c.Key(key), "").Label(label).Tooltip(label).
		Size(28, 28).Padding(0).Radius(7).MinWidth(0)
	b.Children(func() {
		ui.Icon(c, icon).FontSize(fsLead)
	})
	return b.OnClick(on)
}

// pill 是一小块带底色的短文本，例如科目名、日期标记。
//
// 参数 bg、fg 是底色与文字色，通常由调用方用强调色的不同透明度拼出来。
func pill(c *ui.Context, text string, bg, fg ui.Color) ui.Element {
	return ui.Row(c).Padding(2.5, 8).Radius(6).Background(bg).
		AlignItems(ui.Center).Shrink(0).Children(func() {
		ui.Text(c, text).FontSize(fsCaption).FontWeight(600).TextColor(fg).SingleLine()
	})
}

// dot 是条目前的圆点。
//
// 参数 size 是直径（DIP），color 是填充色。
func dot(c *ui.Context, size float32, color ui.Color) ui.Element {
	return ui.Box(c).Size(size, size).Radius(size / 2).Background(color).Shrink(0)
}

// kvRow 是设置页里的一行：左边标签，右边内容。
//
// 参数 label 是固定的短标签，value 构建右侧的内容。
func kvRow(c *ui.Context, label string, value func()) ui.Element {
	t := c.Theme()
	return ui.Row(c).AlignItems(ui.Center).Gap(16).Padding(7, 0).Children(func() {
		ui.Text(c, label).FontSize(fsBody).TextColor(t.TextMuted).Width(96).Shrink(0)
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
