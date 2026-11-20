package pages

import (
	"fmt"
	"strings"

	"flowlog-desktop/internal/app"
	design "flowlog-desktop/internal/design"
	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// Notes 是随笔页：按日期倒序的一列文字卡片，可以搜索。
//
// 随笔不按科目组织、也没有勾选，就是每天随手写下的东西。
func Notes(c *ui.Context, a *app.App) {
	p := design.PaletteOf(c)
	notes := a.FilteredNotes()

	subtitle := fmt.Sprintf("共 %d 条", len(a.Notes))
	if strings.TrimSpace(a.NoteQuery) != "" {
		subtitle = fmt.Sprintf("「%s」· %d 条", strings.TrimSpace(a.NoteQuery), len(notes))
	}

	ui.Column(c).Grow(1).MinWidth(0).Children(func() {
		design.PageHeader(c, "随笔", subtitle, func() {
			search := ui.SearchField(c.Key("note-search"), &a.NoteQuery).
				Placeholder("搜索随笔").Label("搜索随笔").Width(220).
				Height(design.ControlHeight)
			_ = search
			design.FilledButton(c, "note-add", "写随笔", design.IconPen, func() {
				a.OpenNoteEditor(a.Today, "")
			})
		})

		if len(notes) == 0 {
			if strings.TrimSpace(a.NoteQuery) != "" {
				design.EmptyState(c, design.IconSearch, "没有匹配的随笔",
					"换个关键词试试", "清空搜索", func() { a.NoteQuery = "" })
			} else {
				design.EmptyState(c, design.IconNote, "还没有随笔",
					"写下今天真正想记下来的事，不必是学习内容", "写第一条随笔", func() {
						a.OpenNoteEditor(a.Today, "")
					})
			}
			return
		}

		a.NoteList.Key = func(i int) any {
			if i < 0 || i >= len(notes) {
				return i
			}
			return notes[i].ID
		}
		ui.List(c, &a.NoteList, len(notes), func(i int) {
			if i < 0 || i >= len(notes) {
				return
			}
			noteCard(c, a, notes[i])
		}).Grow(1).Padding(2, 32, 26).Gap(14)
		_ = p
	})
}

// noteCard 是一条随笔：日期头、正文，以及悬停时出现的编辑与删除。
func noteCard(c *ui.Context, a *app.App, note model.Note) {
	p := design.PaletteOf(c)
	rel := model.RelativeDay(note.Date)

	card := ui.Column(c.Key("note-"+note.ID)).Padding(20, 22).Gap(10).
		Radius(design.RadiusMD).Background(cardSurface(c)).
		Shadow(0, 1, 2, 0, ui.RGBA(23, 32, 74, 0.04)).
		Shadow(0, 6, 16, 0, ui.RGBA(23, 32, 74, 0.05))
	hovered := card.Hovered()

	card.Children(func() {
		ui.Row(c).Gap(9).AlignItems(ui.Center).Children(func() {
			ui.Text(c, model.DisplayDate(note.Date)).FontSize(design.FontHeading).
				FontWeight(600).SingleLine().TextColor(p.OnSurface)
			ui.Text(c, model.Weekday(model.MustDate(note.Date))).
				FontSize(design.FontSmall).TextColor(p.OnSurfaceVariant)
			if rel != "" {
				design.Chip(c, rel, design.TonePrimary)
			}
			ui.Spacer(c)
			if note.At != "" {
				ui.Text(c, note.At).FontSize(design.FontCaption).TextColor(p.OnSurfaceVariant)
			}
			opacity := 0.45 + 0.55*boolFloat(hovered)
			edit := design.IconButton(c, "note-edit-"+note.ID, "编辑这条随笔", design.IconPen, func() {
				a.OpenNoteEditor(note.Date, note.ID)
			})
			edit.Opacity(card.Animate("note-edit", opacity, design.MotionState.Duration))
			del := design.IconButton(c, "note-del-"+note.ID, "删除这条随笔", design.IconTrash, func() {
				a.AskDeleteNote(note.ID)
			})
			del.Opacity(card.Animate("note-del", opacity, design.MotionState.Duration))
		})

		highlightedText(c, note.Text, a.NoteQuery, design.FontBody, p.OnSurface).
			LineHeight(1.65)
	})
}

// NoteEditor 是写 / 改随笔的弹窗。
func NoteEditor(c *ui.Context, a *app.App) {
	if !a.NoteEditor.Open {
		return
	}
	p := design.PaletteOf(c)
	editing := a.NoteEditor.ID != ""

	title := "写随笔"
	if editing {
		title = "编辑随笔"
	}

	design.Dialog(c, &a.NoteEditor.Open, func() {
		_, windowHeight := c.Size()
		formHeight := windowHeight - 340
		if formHeight < 200 {
			formHeight = 200
		}

		ui.Column(c).Width(560).Gap(16).Children(func() {
			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
				ui.Text(c, title).FontSize(design.FontHeading).FontWeight(600).
					TextColor(p.OnSurface).SingleLine()
				ui.Spacer(c)
				ui.Text(c, app.CmdName+" + Enter 保存").FontSize(design.FontTiny).
					TextColor(p.OnSurfaceVariant)
				design.IconButton(c, "close-note-editor", "关闭", design.IconClose, func() {
					a.NoteEditor.Open = false
				})
			})

			ui.Column(c).Width(168).Children(func() {
				ui.Field(c, "日期", func() {
					ui.DateInput(c.Key("note-date"), &a.NoteEditor.Date).Label("日期")
				})
			})

			ui.Field(c, "内容", func() {
				ui.TextArea(c.Key("note-text"), &a.NoteEditor.Text).Label("内容").
					Lines(8, 16).
					Placeholder("今天想记下来的事…")
			})

			if a.NoteEditor.Err != "" {
				ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Icon(c, design.IconAlert).FontSize(design.FontBody).TextColor(p.Error)
					ui.Text(c, a.NoteEditor.Err).FontSize(design.FontSmall).
						TextColor(p.Error).Grow(1)
				})
			}

			design.Divider(c)
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				if editing {
					design.DangerTextButton(c, "delete-note", "删除这条", func() {
						id := a.NoteEditor.ID
						a.NoteEditor.Open = false
						a.AskDeleteNote(id)
					})
				}
				ui.Spacer(c)
				design.TextButton(c, "cancel-note", "取消", func() { a.NoteEditor.Open = false })
				design.FilledButton(c, "save-note", "保存", nil, a.SaveNoteEditor)
			})
		})

		if c.Shortcut(ui.Cmd, ui.KeyEnter) {
			a.SaveNoteEditor()
		}
	})
}
