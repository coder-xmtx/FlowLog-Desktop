package app

import (
	"strings"
	"time"

	"flowlog-desktop/internal/model"
)

// 随笔的动作：写、改、删、搜索。
//
// 随笔和学记、待办是三套独立的数据；这里只碰 Notes 和数据库的 notes 表。

// NoteEditorState 是「写随笔」弹窗的草稿。
type NoteEditorState struct {
	// Open 表示弹窗是否显示。
	Open bool
	// Date 是这条随笔所属的日期。
	Date time.Time
	// ID 是正在编辑的随笔 ID；空字符串表示新增。
	ID string
	// Text 是正文。
	Text string
	// Err 是校验失败时的提示。
	Err string
}

// OpenNoteEditor 打开随笔编辑器：id 为空表示新增。
func (a *App) OpenNoteEditor(date, id string) {
	ed := &a.NoteEditor
	ed.Open = true
	ed.ID = id
	ed.Err = ""
	ed.Date = model.MustDate(date)
	if ed.Date.IsZero() {
		ed.Date, _ = model.ParseDate(a.Today)
	}
	ed.Text = ""
	if id == "" {
		return
	}
	for _, note := range a.Notes {
		if note.ID == id {
			ed.Text = note.Text
			ed.Date = model.MustDate(note.Date)
			return
		}
	}
}

// SaveNoteEditor 校验草稿并写入。
func (a *App) SaveNoteEditor() {
	a.NoteEditor.Err = ""
	text := strings.TrimSpace(a.NoteEditor.Text)
	if text == "" {
		a.NoteEditor.Err = "写点什么再保存"
		return
	}
	date := a.NoteEditor.Date.Format(model.DateLayout)
	note := model.Note{
		ID:   a.NoteEditor.ID,
		Date: date,
		Text: text,
		At:   time.Now().Format("15:04"),
	}
	if note.ID == "" {
		note.ID = model.NewID()
	}
	if err := a.Store.SaveNote(note); err != nil {
		a.NoteEditor.Err = err.Error()
		return
	}
	if i := a.noteIndex(note.ID); i >= 0 {
		a.Notes[i] = note
	} else {
		a.Notes = append(a.Notes, note)
	}
	a.sortNotes()
	a.NoteEditor.Open = false
	a.Toast("已保存", "随笔 · "+model.DisplayDate(date))
}

// noteIndex 返回某条随笔在 Notes 里的下标。
func (a *App) noteIndex(id string) int {
	for i := range a.Notes {
		if a.Notes[i].ID == id {
			return i
		}
	}
	return -1
}

// sortNotes 保持与数据库一致的顺序：日期倒序，同一天按写入时刻倒序。
func (a *App) sortNotes() {
	// 切片很小（一年最多几百条），插入排序足够且稳定。
	for i := 1; i < len(a.Notes); i++ {
		for j := i; j > 0; j-- {
			x, y := a.Notes[j-1], a.Notes[j]
			before := x.Date > y.Date || (x.Date == y.Date && x.At > y.At)
			if !before {
				break
			}
			a.Notes[j-1], a.Notes[j] = y, x
		}
	}
}

// AskDeleteNote 弹出确认，删掉一条随笔。
func (a *App) AskDeleteNote(id string) {
	i := a.noteIndex(id)
	if i < 0 {
		return
	}
	text := firstLine(a.Notes[i].Text)
	a.Confirm = ConfirmState{
		Open:    true,
		Title:   "删除这条随笔？",
		Message: "「" + text + "」，删除后无法恢复。",
		Yes:     "删除",
		YesFunc: func() { a.DeleteNote(id) },
	}
}

// DeleteNote 删除一条随笔，并给一次撤销的机会。
func (a *App) DeleteNote(id string) {
	i := a.noteIndex(id)
	if i < 0 {
		return
	}
	removed := a.Notes[i]
	if err := a.Store.DeleteNote(id); err != nil {
		a.ToastError("删除随笔失败", err.Error())
		return
	}
	a.Notes = append(a.Notes[:i], a.Notes[i+1:]...)
	a.ToastUndo("已删除随笔", "撤销", func() {
		if err := a.Store.SaveNote(removed); err != nil {
			a.ToastError("恢复随笔失败", err.Error())
			return
		}
		a.Notes = append(a.Notes, removed)
		a.sortNotes()
	})
}

// FilteredNotes 按搜索词筛选随笔，日期倒序。
//
// 空搜索返回全部；命中的文字由页面负责高亮。
func (a *App) FilteredNotes() []model.Note {
	q := strings.TrimSpace(a.NoteQuery)
	if q == "" {
		return a.Notes
	}
	out := make([]model.Note, 0, len(a.Notes))
	for _, note := range a.Notes {
		if model.QueryHit(note.Text, q) || model.QueryHit(model.DisplayDate(note.Date), q) {
			out = append(out, note)
		}
	}
	return out
}

// firstLine 取多行文本的第一行，用于确认框里的一句话摘要。
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

// TodayNoteCount 返回今天写了几条随笔。
func (a *App) TodayNoteCount() int {
	n := 0
	for _, note := range a.Notes {
		if note.Date == a.Today {
			n++
		}
	}
	return n
}
