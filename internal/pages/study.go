package pages

import (
	"fmt"
	"strings"

	"flowlog-desktop/internal/app"
	design "flowlog-desktop/internal/design"
	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// allSubjects 是科目筛选里「不限科目」那一项的文字。
const allSubjects = "全部科目"

// Study 是学记页：按日期倒序的记录卡片，右上角写新的一条。
//
// 页头右侧有科目筛选与搜索框；两者可以叠加，卡片里只显示命中的记录，
// 命中的文字用琥珀色标出来。
func Study(c *ui.Context, a *app.App) {
	p := design.PaletteOf(c)
	a.FilterDays()

	subtitle := fmt.Sprintf("共 %d 天 · %d 条要点", a.DaysWritten(), a.TotalPoints())
	if a.HasFilter() {
		subtitle = a.FilterSummary()
	}

	ui.Column(c).Grow(1).MinWidth(0).Children(func() {
		design.PageHeader(c, "学记", subtitle, func() {
			subjectSelect(c, a)
			searchField(c, a)
			design.FilledButton(c, "study-add", "写学记", design.IconPlus, func() {
				a.OpenEditor(a.Today, "")
			})
		})

		if len(a.Filtered) == 0 {
			if a.HasFilter() {
				design.EmptyState(c, design.IconSearch, "没有匹配的记录",
					"换个关键词，或者切换到别的科目试试。", "清空筛选", a.ClearFilters)
			} else {
				design.EmptyState(c, design.IconBook, "还没有记录",
					"写下科目、章节，以及今天弄懂的要点。", "写今天的学记", func() {
						a.OpenEditor(a.Today, "")
					})
			}
			return
		}

		// 列表按日期给每一行一个稳定 Key：滚动位置与行内控件状态跟着
		// 「那一天」走，而不是跟着下标走。
		a.StudyList.Key = func(i int) any {
			if i < 0 || i >= len(a.Filtered) {
				return i
			}
			return a.Days[a.Filtered[i]].Date
		}
		ui.List(c, &a.StudyList, len(a.Filtered), func(i int) {
			dayCard(c, a, a.Filtered[i])
		}).Grow(1).Padding(2, 32, 26).Gap(16)

		// 侧边栏日历点了一天：滚到那一天。只滚一次。
		if a.RevealDate != "" {
			for i, di := range a.Filtered {
				if a.Days[di].Date == a.RevealDate {
					a.StudyList.ScrollIntoView(i)
					break
				}
			}
			a.RevealDate = ""
		}
		_ = p
	})
}

// subjectSelect 是页头右侧的科目筛选下拉框。
//
// 选项来自全部记录里出现过的科目，按出现次数从多到少排列；
// 只有存在科目时才显示。
func subjectSelect(c *ui.Context, a *app.App) {
	subjects := a.SubjectCounts()
	if len(subjects) == 0 {
		return
	}
	options := make([]string, 0, len(subjects)+1)
	options = append(options, allSubjects)
	for _, s := range subjects {
		options = append(options, s.Name)
	}

	// Select 绑定的是「选中的字符串」。本地副本每帧由 SubjectFilter 推出，
	// 用户选中后由 OnChange 写回，避免直接绑派生值。
	choice := a.SubjectFilter
	if choice == "" {
		choice = allSubjects
	}
	ui.Select(c.Key("study-subject"), &choice, options).
		Label("按科目筛选").Width(168).Tooltip("只看某个科目").
		OnChange(func() {
			if choice == allSubjects {
				a.SubjectFilter = ""
				return
			}
			a.SubjectFilter = choice
		})
}

// searchField 是页头右侧的搜索框，⌘F 把焦点送进来。
func searchField(c *ui.Context, a *app.App) {
	f := ui.SearchField(c.Key("study-search"), &a.Query).
		Placeholder("搜索科目、章节或要点").Label("搜索学记").Width(210)
	if a.FocusSearch {
		f.Focus()
		a.FocusSearch = false
	}
}

// dayCard 画一天：日期头与若干条学习记录。
//
// 参数 di 是这一日在 Days 里的下标。卡片里只画当前筛选下可见的条目，
// 所以按科目筛选时，一张卡片里不会出现别的科目。
func dayCard(c *ui.Context, a *app.App, di int) {
	p := design.PaletteOf(c)
	if di < 0 || di >= len(a.Days) {
		return
	}
	day := a.Days[di]
	date := day.Date
	rel := model.RelativeDay(date)

	shown := a.ShownEntries(day)
	points := 0
	for _, ei := range shown {
		points += len(day.Entries[ei].Points)
	}

	card := ui.Column(c.Key("day-card-"+date)).Padding(20, 22).Gap(14).
		Radius(design.RadiusMD).Background(cardSurface(c)).
		Shadow(0, 1, 2, 0, ui.RGBA(23, 32, 74, 0.04)).
		Shadow(0, 6, 16, 0, ui.RGBA(23, 32, 74, 0.05))
	hovered := card.Hovered()

	card.Children(func() {
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			ui.Text(c, model.DisplayDate(date)).FontSize(design.FontHeading).
				FontWeight(600).SingleLine().TextColor(p.OnSurface)
			ui.Text(c, model.Weekday(model.MustDate(date))).FontSize(design.FontSmall).
				TextColor(p.OnSurfaceVariant)
			if rel != "" {
				design.Chip(c, rel, design.TonePrimary)
			}
			if points > 0 {
				text := plural(points, "条要点")
				if done := day.DonePoints(); done > 0 {
					text += " · 已复习 " + fmt.Sprint(done)
				}
				ui.Text(c, text).FontSize(design.FontCaption).TextColor(p.OnSurfaceVariant)
			}
			ui.Spacer(c)
			// 删除整天：平时半透明，鼠标移到卡片上才完全显示。
			del := design.IconButton(c, "del-"+date, "删除这一天的记录", design.IconTrash, func() {
				a.AskDeleteDay(date)
			})
			del.Opacity(card.Animate("del", boolFloat(hovered), design.MotionState.Duration))
		})

		if len(day.Entries) == 0 {
			ui.Text(c, "这一天还没有记录。").FontSize(design.FontBody).
				TextColor(p.OnSurfaceVariant).Padding(2, 0)
			return
		}

		for n, ei := range shown {
			if n > 0 {
				design.Divider(c)
			}
			entryBlock(c, a, di, ei)
		}
	})
}

// entryBlock 是一条学习记录：科目、章节路径、要点清单、备注。
//
// 参数 di、ei 分别是这一日在 Days 里的下标和条目在当天的下标。
// 有搜索词时，命中的文字带琥珀底色。
func entryBlock(c *ui.Context, a *app.App, di, ei int) {
	p := design.PaletteOf(c)
	e := &a.Days[di].Entries[ei]
	date := a.Days[di].Date
	id := e.ID
	q := a.Query

	entry := ui.Column(c.Key("entry-" + id)).Gap(9)
	hovered := entry.Hovered()

	entry.Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			design.Chip(c, e.Subject, design.TonePrimary)

			if topic := design.TopicLine(*e); topic != "" {
				highlightedText(c, topic, q, design.FontSmall, p.OnSurfaceVariant).
					SingleLine().Grow(1)
			} else {
				ui.Spacer(c)
			}

			if e.Minutes > 0 {
				ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
					ui.Icon(c, design.IconClock).FontSize(design.FontCaption).
						TextColor(p.OnSurfaceVariant)
					ui.Text(c, design.FormatMinutes(e.Minutes)).FontSize(design.FontCaption).
						TextColor(p.OnSurfaceVariant)
				})
			}
			if e.At != "" {
				ui.Text(c, e.At).FontSize(design.FontCaption).TextColor(p.OnSurfaceVariant)
			}

			// 编辑按钮平时半透明，鼠标移到这一条上才变亮。
			edit := design.IconButton(c, "edit-entry-"+id, "编辑这条记录", design.IconPen, func() {
				a.OpenEditor(date, id)
			})
			edit.Opacity(entry.Animate("edit", 0.45+0.55*boolFloat(hovered), design.MotionState.Duration))
		})

		if len(e.Points) > 0 {
			ui.Column(c).Gap(6).Padding(0, 2).Children(func() {
				for k := range e.Points {
					pointRow(c, a, di, ei, k)
				}
			})
		}

		if strings.TrimSpace(e.Note) != "" {
			// 备注放在一块浅色容器里，和要点分开，读起来更轻。
			ui.Row(c).Gap(10).Padding(10, 12).Radius(design.RadiusSM).
				Background(p.SurfaceContainer).Children(func() {
				highlightedText(c, e.Note, q, design.FontBody, p.OnSurfaceVariant).
					LineHeight(1.55).Grow(1)
			})
		}
	})
}

// pointRow 是一条要点：勾上表示已经复习。
//
// 参数 di、ei、pi 分别是日、条目、要点在各自切片里的下标。
// 勾选直接写进 model，OnChange 再负责落库。
func pointRow(c *ui.Context, a *app.App, di, ei, pi int) {
	p := design.PaletteOf(c)
	pt := &a.Days[di].Entries[ei].Points[pi]
	date := a.Days[di].Date

	ui.Row(c).Gap(10).AlignItems(ui.Start).Children(func() {
		cb := design.Checkbox(c, "pt-"+pt.ID, &pt.Done, pt.Text).
			Margin(design.CheckboxOffset(), 0, 0, 0)
		cb.Tooltip("标记为已复习")
		cb.OnChange(func() { a.Persist(date) })

		color := p.OnSurface
		if pt.Done {
			color = p.OnSurfaceVariant
		}
		text := highlightedText(c, pt.Text, a.Query, design.FontBody, color).
			LineHeight(1.45).Grow(1)
		if pt.Done {
			text.Strikethrough()
		}
	})
}

// boolFloat 把布尔值转成动画目标 0 或 1。
func boolFloat(b bool) float32 {
	if b {
		return 1
	}
	return 0
}

// noteSurface 返回备注块的底色：深色下比卡片再亮一层，浅色下是浅灰容器。
func noteSurface(c *ui.Context) ui.Color {
	p := design.PaletteOf(c)
	if p.Dark {
		return p.SurfaceContainerHigh
	}
	return p.SurfaceContainer
}
