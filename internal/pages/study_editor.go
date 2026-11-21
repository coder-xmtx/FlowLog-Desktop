package pages

import (
	"flowlog-desktop/internal/app"
	design "flowlog-desktop/internal/design"
	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// 写学记的弹窗：科目、章、节、要点与备注。
//
// 弹窗本身用 Material 3 的对话框样式；表单区域可滚动，错误提示与按钮固定在底部。
func StudyEditor(c *ui.Context, a *app.App) {
	if !a.Editor.Open {
		return
	}
	p := design.PaletteOf(c)
	editing := a.Editor.EntryID != ""

	title := "写今天的学记"
	switch {
	case editing:
		title = "编辑这条记录"
	case a.Editor.Date.Format(model.DateLayout) != a.Today:
		title = "补记 " + model.DisplayDate(a.Editor.Date.Format(model.DateLayout))
	}

	design.Dialog(c, &a.Editor.Open, func() {
		// 表单高度跟着窗口走：窗口矮时表单自己滚动，按钮始终留在底部。
		_, windowHeight := c.Size()
		formHeight := windowHeight - 360
		if formHeight < 180 {
			formHeight = 180
		}

		ui.Column(c).Width(560).Gap(16).Children(func() {
			// 标题行
			ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
				design.Text(c, title).FontSize(design.FontHeading).FontWeight(600).
					TextColor(p.OnSurface).SingleLine()
				ui.Spacer(c)
				design.Text(c, app.CmdName+design.T(" + Enter 保存")).FontSize(design.FontTiny).
					TextColor(p.OnSurfaceVariant)
				design.IconButton(c, "close-editor", "关闭", design.IconClose, func() {
					a.Editor.Open = false
				})
			})

			ui.Scroll(c).Scrollbars(false).MaxHeight(formHeight).Children(func() {
				ui.Column(c).Gap(14).Children(func() {
					// 日期 + 科目
					ui.Row(c).Gap(12).AlignItems(ui.Start).Children(func() {
						if editing {
							ui.Column(c).Width(168).Gap(6).Children(func() {
								ui.Row(c).Gap(5).AlignItems(ui.Center).Children(func() {
									ui.Icon(c, design.IconCalendar).FontSize(design.FontCaption).
										TextColor(p.OnSurfaceVariant)
									design.Text(c, "日期").FontSize(design.FontCaption).
										TextColor(p.OnSurfaceVariant)
								})
								ui.Text(c, model.DisplayDate(a.Editor.Date.Format(model.DateLayout))).
									FontSize(design.FontBody).Padding(7, 0)
							})
						} else {
							ui.Column(c).Width(168).Children(func() {
								ui.Field(c, "日期", func() {
									ui.DateInput(c.Key("editor-date"), &a.Editor.Date).Label(design.T("日期"))
								})
							})
						}
						ui.Column(c).Grow(1).MinWidth(0).Children(func() {
							ui.Field(c, "科目", func() {
								ui.Autocomplete(c.Key("editor-subject"), &a.Editor.Subject, a.Subjects()).
									Placeholder("计算机组成原理").Label(design.T("科目")).AutoFocus()
							})
						})
					})

					// 章、节分成两个可选输入框
					ui.Row(c).Gap(12).AlignItems(ui.Start).Children(func() {
						ui.Column(c).Grow(1).MinWidth(0).Children(func() {
							ui.Field(c, "章", func() {
								ui.TextInput(c.Key("editor-chapter"), &a.Editor.Chapter).
									Label("章").Placeholder(design.T("第 4 章 存储系统"))
							})
						})
						ui.Column(c).Grow(1).MinWidth(0).Children(func() {
							ui.Field(c, "节", func() {
								ui.TextInput(c.Key("editor-section"), &a.Editor.Section).
									Label("节").Placeholder(design.T("第 8 节 虚拟存储器"))
							})
						})
					})
					design.Text(c, "章、节都可以留空，只写科目也能保存").
						FontSize(design.FontTiny).TextColor(p.OnSurfaceVariant)

					ui.Field(c, "要点", func() {
						ui.TextArea(c.Key("editor-points"), &a.Editor.Points).Label(design.T("要点")).
							Lines(5, 10).
							ScrollbarAxes(false, false).
							Placeholder("一行一条，例如：\n页式虚拟存储器的地址映射\n页表的结构与访问流程")
					})
					design.Text(c, "每行一条，会自动去掉开头的「- 」").
						FontSize(design.FontTiny).TextColor(p.OnSurfaceVariant)

					// 时长 + 备注
					ui.Row(c).Gap(12).AlignItems(ui.Start).Children(func() {
						ui.Column(c).Width(168).Children(func() {
							ui.Field(c, "学习时长（分钟）", func() {
								design.NumberField(c, "editor-minutes", &a.Editor.Minutes, 0, 600, 5, 56, nil)
							})
						})
						ui.Column(c).Grow(1).MinWidth(0).Children(func() {
							ui.Field(c, "备注", func() {
								ui.TextArea(c.Key("editor-note"), &a.Editor.Note).Label(design.T("备注")).
									Lines(1, 3).
									ScrollbarAxes(false, false).
									Placeholder("可选")
							})
						})
					})
				})
			})

			// 校验错误放在滚动区外面：不管表单滚到哪里，都能立刻看到。
			if a.Editor.Err != "" {
				ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Icon(c, design.IconAlert).FontSize(design.FontBody).TextColor(p.Error)
					ui.Text(c, a.Editor.Err).FontSize(design.FontSmall).TextColor(p.Error).Grow(1)
				})
			}

			design.Divider(c)
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				if editing {
					design.DangerTextButton(c, "delete-entry", "删除这条", func() {
						date := a.Editor.Date.Format(model.DateLayout)
						id := a.Editor.EntryID
						a.Editor.Open = false
						a.AskDeleteEntry(date, id)
					})
				}
				ui.Spacer(c)
				design.TextButton(c, "cancel-editor", "取消", func() { a.Editor.Open = false })
				design.FilledButton(c, "save-editor", "保存", nil, a.SaveEditor)
			})
		})

		if c.Shortcut(ui.Cmd, ui.KeyEnter) {
			a.SaveEditor()
		}
	})
}
