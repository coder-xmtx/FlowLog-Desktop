package design

import (
	"fmt"
	"strings"

	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// PaletteOf 返回当前帧对应的颜色板。
//
// 颜色板的深浅由主题的 Dark 决定，和我们组装主题时用的是同一份数据，
// 所以组件里取角色色时直接用它，不必再把颜色板一路传参。
func PaletteOf(c *ui.Context) Palette { return For(c.Theme().Dark) }

// Card 是一张 Material 3 的悬浮卡片。
//
// 浅色下是白底 + 极淡投影，深色下是略高于背景的容器色，两者都不描边，
// 靠层次而不是边框区分，观感更轻。参数 body 在卡片内构建内容。
func Card(c *ui.Context, body func()) ui.Element {
	p := PaletteOf(c)
	e := ui.Column(c).Padding(20).Gap(14).Radius(RadiusMD)
	if p.Dark {
		e = e.Background(p.SurfaceContainer)
	} else {
		e = e.Background(p.SurfaceContainerLowest).
			Shadow(0, 1, 2, 0, ui.RGBA(23, 32, 74, 0.04)).
			Shadow(0, 6, 16, 0, ui.RGBA(23, 32, 74, 0.05))
	}
	return e.Children(body)
}

// PageHeader 是页面顶部的标题区：左边标题与副标题，右边操作。
//
// 参数 title 是页面名，subtitle 是一行说明（可为空），actions 构建右侧操作。
// 窗口窄时（内容区放不下两侧）自动改成上下两行，标题不会被挤没。
func PageHeader(c *ui.Context, title, subtitle string, actions func()) ui.Element {
	p := PaletteOf(c)
	w, _ := c.Size()
	stacked := actions != nil && w < 1020

	titleBlock := func() {
		ui.Column(c).Gap(4).Grow(1).MinWidth(0).Children(func() {
			ui.Text(c, title).FontSize(FontTitle).FontWeight(600).LetterSpacing(-0.3).
				SingleLine().TextColor(p.OnSurface)
			if subtitle != "" {
				ui.Text(c, subtitle).FontSize(FontBody).TextColor(p.OnSurfaceVariant).SingleLine()
			}
		})
	}

	if stacked {
		return ui.Column(c).Gap(12).Padding(24, 28, 14).Children(func() {
			ui.Row(c).Children(titleBlock)
			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(actions)
		})
	}

	return ui.Row(c).AlignItems(ui.Center).Gap(16).Padding(28, 32, 16).Children(func() {
		titleBlock()
		if actions != nil {
			actions()
		}
	})
}

// SectionCard 是带小标题的卡片，设置页与首页的区块都用它。
//
// 参数 title 是区块标题，hint 是标题右侧的浅色补充（可为空），body 构建内容。
func SectionCard(c *ui.Context, title, hint string, body func()) ui.Element {
	p := PaletteOf(c)
	return Card(c, func() {
		ui.Row(c).AlignItems(ui.Center).Gap(8).Children(func() {
			ui.Text(c, title).FontSize(FontLead).FontWeight(600).TextColor(p.OnSurface)
			if hint != "" {
				ui.Text(c, hint).FontSize(FontCaption).TextColor(p.OnSurfaceVariant)
			}
		})
		body()
	})
}

// Dialog 是 Material 3 的弹窗：半透明遮罩 + 高一层容器色的面板 + 28 圆角。
//
// 参数 open 控制显示，body 在面板里构建内容。点遮罩或按 Esc 会把它设为 false。
func Dialog(c *ui.Context, open *bool, body func()) ui.Element {
	p := PaletteOf(c)
	return ui.DialogBase(c, open, func(back, panel ui.Element) {
		back.Background(ui.RGBA(6, 10, 28, 0.42))
		panel.Padding(24).Gap(18).Radius(RadiusLG).
			Background(p.SurfaceContainerHigh).MaxWidth(720).
			Shadow(0, 24, 48, 0, ui.RGBA(6, 10, 28, 0.24))
		panel.Children(body)
	})
}

// Confirm 是一个 Material 3 的确认弹窗。
//
// 参数 open 控制显示，title 与 message 是标题与正文，yes 是确认按钮文字，
// onYes 在点确认时执行；取消、Esc、点遮罩都只是关掉它。
func Confirm(c *ui.Context, open *bool, title, message, yes string, onYes func()) ui.Element {
	p := PaletteOf(c)
	return Dialog(c, open, func() {
		ui.Text(c, title).FontSize(FontHeading).FontWeight(600).TextColor(p.OnSurface)
		ui.Text(c, message).FontSize(FontBody).TextColor(p.OnSurfaceVariant).LineHeight(1.55)
		ui.Row(c).Gap(8).AlignItems(ui.Center).Padding(4, 0, 0).Children(func() {
			ui.Spacer(c)
			TextButton(c, "confirm-cancel", "取消", func() { *open = false })
			FilledButton(c, "confirm-yes", yes, nil, func() {
				*open = false
				if onYes != nil {
					onYes()
				}
			})
		})
	})
}

// Divider 是一条 1 DIP 的分隔线，颜色用最弱的描边色。
func Divider(c *ui.Context) ui.Element {
	return ui.Box(c).Height(1).Background(PaletteOf(c).OutlineVariant).Margin(2, 0)
}

// Tone 是芯片的语气：决定底色与文字色。
type Tone int

const (
	// ToneNeutral 是中性芯片：弱底、次要文字。
	ToneNeutral Tone = iota
	// TonePrimary 是主色芯片：用于科目这类需要认出来的标签。
	TonePrimary
	// ToneSecondary 是辅色芯片：用于日期、数量。
	ToneSecondary
	// ToneTertiary 是第三色芯片：用于「已复习」「已完成」这类正向状态。
	ToneTertiary
	// ToneError 是错误芯片：过期、失败。
	ToneError
)

// Chip 是一枚 Material 3 的标签芯片。
//
// 参数 text 是内容，tone 决定配色；返回的芯片是固定高度的胶囊，
// 不会因为父容器的拉伸而变形。
func Chip(c *ui.Context, text string, tone Tone) ui.Element {
	p := PaletteOf(c)
	bg, fg := chipColors(p, tone)
	return ui.Row(c).Padding(3, 10).Radius(RadiusXS).Background(bg).
		AlignItems(ui.Center).Shrink(0).Children(func() {
		ui.Text(c, text).FontSize(FontCaption).FontWeight(500).TextColor(fg).SingleLine()
	})
}

// chipColors 返回某个语气对应的底色与文字色。
func chipColors(p Palette, tone Tone) (bg, fg ui.Color) {
	switch tone {
	case TonePrimary:
		return p.PrimaryContainer, p.OnPrimaryContainer
	case ToneSecondary:
		return p.SecondaryContainer, p.OnSecondaryContainer
	case ToneTertiary:
		return p.TertiaryContainer, p.OnTertiaryContainer
	case ToneError:
		return p.ErrorContainer, p.OnErrorContainer
	default:
		return p.SurfaceContainerHigh, p.OnSurfaceVariant
	}
}

// EmptyState 是空列表时的引导：图标、说明和一颗按钮。
//
// 参数 icon 是顶部图标，title 与 desc 是两行说明，action 是按钮文字
// （空字符串表示不显示按钮），on 是按钮动作。
func EmptyState(c *ui.Context, icon *ui.SVG, title, desc, action string, on func()) ui.Element {
	p := PaletteOf(c)
	return ui.Column(c).FillWidth().Center().Padding(12, 16).Gap(12).Children(func() {
		ui.Box(c).Size(56, 56).Radius(28).Background(p.SurfaceContainerHigh).
			Center().Children(func() {
			ui.Icon(c, icon).FontSize(24).TextColor(p.OnSurfaceVariant)
		})
		ui.Text(c, title).FontSize(FontLead).FontWeight(600).TextColor(p.OnSurface)
		if desc != "" {
			ui.Text(c, desc).FontSize(FontBody).TextColor(p.OnSurfaceVariant).
				TextAlign(ui.Center).MaxWidth(380).LineHeight(1.5)
		}
		if action != "" && on != nil {
			TonalButton(c, "empty-action", action, nil, on)
		}
	})
}

// 勾选框的尺寸常量，用来把它和旁边的文字对齐。
const (
	// checkboxSize 是 MyGo 勾选框方框的边长（主题 Space(4)）。
	checkboxSize = 16
	// checkboxNudge 是勾选框的光学微调：中文的墨迹中心比行盒中心低约 1 DIP。
	checkboxNudge = 1
	// pointCheckboxOffset 是要点行里勾选框的下移量，让方框居中于第一行文字。
	pointCheckboxOffset = (FontBody*1.45 - checkboxSize) / 2
)

// Checkbox 构建一个和右侧文字视觉对齐的勾选框。
//
// 参数 key 是它在当前列表里的稳定身份，checked 是绑定状态，
// label 是给读屏与自动化用、界面上不显示的文字。
func Checkbox(c *ui.Context, key any, checked *bool, label string) ui.Element {
	return ui.Checkbox(c.Key(key), checked, "").Label(label).
		Translate(0, checkboxNudge)
}

// CheckboxOffset 返回要点行里勾选框需要的下移量。
func CheckboxOffset() float32 { return pointCheckboxOffset }

// IconButton 是一个只有图标的方形按钮。
//
// 参数 key 是稳定身份，label 同时作为无障碍名称与提示，on 是点击动作。
// 悬停时浮现一层强调色的 state layer，带 150ms 过渡。
func IconButton(c *ui.Context, key any, label string, icon *ui.SVG, on func()) ui.Element {
	p := PaletteOf(c)
	b := ui.ButtonBase(c.Key(key)).Label(label).Tooltip(label).
		Width(36).Height(36).Radius(18).Center().Shrink(0)
	hovered := b.Hovered()
	b.Children(func() {
		ui.Icon(c, icon).FontSize(FontLead).TextColor(p.OnSurfaceVariant)
	})
	b.Background(HoverLayer(b, "hover", p.OnSurface, hovered))
	return b.OnClick(on)
}

// FilledButton 是 MD3 的主要按钮：主色实底、胶囊形。
//
// 参数 key 是稳定身份，label 是文字，icon 可为 nil，on 是点击动作。
func FilledButton(c *ui.Context, key any, label string, icon *ui.SVG, on func()) ui.Element {
	p := PaletteOf(c)
	b := ui.ButtonBase(c.Key(key)).Label(label).Height(40).Padding(0, 20).
		Radius(RadiusFull).Gap(8).Shrink(0)
	hovered := b.Hovered()
	b.Children(func() {
		if icon != nil {
			ui.Icon(c, icon).FontSize(FontLead).TextColor(p.OnPrimary)
		}
		ui.Text(c, label).FontSize(FontBody).FontWeight(600).TextColor(p.OnPrimary).SingleLine()
	})
	bg := p.Primary
	if hovered {
		bg = p.Primary.Mix(p.OnPrimary, 0.10)
	}
	b.Background(b.AnimateColor("bg", bg, MotionState))
	return b.OnClick(on)
}

// TonalButton 是 MD3 的次级按钮：辅色容器底、主色文字。
//
// 参数含义同 FilledButton。
func TonalButton(c *ui.Context, key any, label string, icon *ui.SVG, on func()) ui.Element {
	p := PaletteOf(c)
	b := ui.ButtonBase(c.Key(key)).Label(label).Height(40).Padding(0, 18).
		Radius(RadiusFull).Gap(8).Shrink(0)
	hovered := b.Hovered()
	b.Children(func() {
		if icon != nil {
			ui.Icon(c, icon).FontSize(FontLead).TextColor(p.OnSecondaryContainer)
		}
		ui.Text(c, label).FontSize(FontBody).FontWeight(600).TextColor(p.OnSecondaryContainer).
			SingleLine()
	})
	bg := p.SecondaryContainer
	if hovered {
		bg = p.SecondaryContainer.Mix(p.OnSecondaryContainer, 0.08)
	}
	b.Background(b.AnimateColor("bg", bg, MotionState))
	return b.OnClick(on)
}

// TextButton 是 MD3 的文字按钮：无底色，只有文字与 hover 的 state layer。
func TextButton(c *ui.Context, key any, label string, on func()) ui.Element {
	p := PaletteOf(c)
	b := ui.ButtonBase(c.Key(key)).Label(label).Height(36).Padding(0, 14).
		Radius(RadiusFull).Shrink(0)
	hovered := b.Hovered()
	b.Children(func() {
		ui.Text(c, label).FontSize(FontSmall).FontWeight(600).TextColor(p.Primary).SingleLine()
	})
	b.Background(HoverLayer(b, "hover", p.Primary, hovered))
	return b.OnClick(on)
}

// DangerTextButton 是红色的文字按钮，用于「删除」这类破坏性动作。
func DangerTextButton(c *ui.Context, key any, label string, on func()) ui.Element {
	p := PaletteOf(c)
	b := ui.ButtonBase(c.Key(key)).Label(label).Height(36).Padding(0, 14).
		Radius(RadiusFull).Shrink(0)
	hovered := b.Hovered()
	b.Children(func() {
		ui.Text(c, label).FontSize(FontSmall).FontWeight(600).TextColor(p.Error).SingleLine()
	})
	b.Background(HoverLayer(b, "hover", p.Error, hovered))
	return b.OnClick(on)
}

// OutlinedButton 是 MD3 的描边按钮，用于次要但需要边框的动作。
func OutlinedButton(c *ui.Context, key any, label string, icon *ui.SVG, on func()) ui.Element {
	p := PaletteOf(c)
	b := ui.ButtonBase(c.Key(key)).Label(label).Height(40).Padding(0, 18).
		Radius(RadiusFull).Gap(8).Shrink(0).Border(1, p.Outline)
	hovered := b.Hovered()
	b.Children(func() {
		if icon != nil {
			ui.Icon(c, icon).FontSize(FontLead).TextColor(p.Primary)
		}
		ui.Text(c, label).FontSize(FontBody).FontWeight(600).TextColor(p.Primary).SingleLine()
	})
	b.Background(HoverLayer(b, "hover", p.Primary, hovered))
	return b.OnClick(on)
}

// KVRow 是设置页里的一行：左边固定宽度的标签，右边内容。
//
// 参数 label 是短标签，value 构建右侧内容。
func KVRow(c *ui.Context, label string, value func()) ui.Element {
	p := PaletteOf(c)
	return ui.Row(c).AlignItems(ui.Center).Gap(16).Padding(8, 0).Children(func() {
		ui.Text(c, label).FontSize(FontBody).TextColor(p.OnSurfaceVariant).Width(96).Shrink(0)
		value()
	})
}

// FormatMinutes 把分钟数写成「45 分钟」「1 小时 20 分」。
func FormatMinutes(m int) string {
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

// TopicLine 把章、节拼成一行面包屑文本。
func TopicLine(e model.Entry) string {
	return strings.Join(e.TopicParts(), "  ›  ")
}

// FirstLine 取多行文本的第一行，用于列表里的一行摘要。
func FirstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}
