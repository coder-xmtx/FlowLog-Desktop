package app

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"flowlog-desktop/internal/model"
)

// 学记（原来的「日记」）的动作：筛选、编辑、删除与撤销。
//
// 页面只负责把这些状态画出来；这里的每个方法都会顺手处理落库与提示。

// HasFilter 报告当前有没有筛选条件。
func (a *App) HasFilter() bool {
	return a.SubjectFilter != "" || strings.TrimSpace(a.Query) != ""
}

// ClearFilters 清空科目筛选与搜索词，回到全部记录。
func (a *App) ClearFilters() {
	a.SubjectFilter = ""
	a.Query = ""
	a.FocusSearch = false
}

// TotalPoints 返回全部学记里的要点条数。
func (a *App) TotalPoints() int {
	n := 0
	for i := range a.Days {
		n += a.Days[i].PointCount()
	}
	return n
}

// FilterSummary 是学记页的副标题：说明当前在看什么。
//
// 有科目筛选时用「科目 · N 天 · M 条要点」，有搜索词时再加上「关键词」。
func (a *App) FilterSummary() string {
	points := 0
	for _, di := range a.Filtered {
		for _, ei := range a.ShownEntries(a.Days[di]) {
			points += len(a.Days[di].Entries[ei].Points)
		}
	}
	q := strings.TrimSpace(a.Query)
	switch {
	case a.SubjectFilter != "" && q != "":
		return fmt.Sprintf("%s ·「%s」· %d 天 · %d 条要点", a.SubjectFilter, q, len(a.Filtered), points)
	case a.SubjectFilter != "":
		return fmt.Sprintf("%s · %d 天 · %d 条要点", a.SubjectFilter, len(a.Filtered), points)
	default:
		return fmt.Sprintf("「%s」· %d 天 · %d 条要点", q, len(a.Filtered), points)
	}
}

// SubjectStat 是一个科目和它在全部记录里出现的次数。
type SubjectStat struct {
	// Name 是科目名。
	Name string
	// Count 是出现次数。
	Count int
}

// SubjectCounts 统计每个科目出现过多少次，按次数从多到少、同次数按名字排。
func (a *App) SubjectCounts() []SubjectStat {
	counts := make(map[string]int, 8)
	for _, d := range a.Days {
		for _, e := range d.Entries {
			if e.Subject == "" {
				continue
			}
			counts[e.Subject]++
		}
	}
	out := make([]SubjectStat, 0, len(counts))
	for name, n := range counts {
		out = append(out, SubjectStat{Name: name, Count: n})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Count != out[j].Count {
			return out[i].Count > out[j].Count
		}
		return out[i].Name < out[j].Name
	})
	return out
}

// FilterDays 按科目筛选与搜索词算出要显示的日子。
//
// 返回 Days 的下标；没有筛选时是全部日子（今天永远在），有筛选时只留下
// 至少有一条可见记录的日子。
func (a *App) FilterDays() []int {
	a.Filtered = a.Filtered[:0]
	for i := range a.Days {
		d := a.Days[i]
		if !a.HasFilter() {
			a.Filtered = append(a.Filtered, i)
			continue
		}
		if d.Date == a.Today && !d.HasEntries() {
			continue
		}
		if len(a.ShownEntries(d)) > 0 {
			a.Filtered = append(a.Filtered, i)
		}
	}
	return a.Filtered
}

// ShownEntries 返回这一天在当前筛选下要显示的条目下标。
//
// 没有筛选时是全部条目；有科目筛选或搜索词时只留下命中的条目，
// 这样一张卡片里不会混着「为什么要显示它」看不出来的记录。
func (a *App) ShownEntries(day model.Day) []int {
	if !a.HasFilter() {
		idx := make([]int, len(day.Entries))
		for i := range idx {
			idx[i] = i
		}
		return idx
	}
	out := make([]int, 0, len(day.Entries))
	for i, e := range day.Entries {
		if a.EntryVisible(e) {
			out = append(out, i)
		}
	}
	return out
}

// EntryVisible 报告一条记录是否满足当前的科目筛选与搜索词。
//
// 科目必须完全相等（下拉框选出来的），搜索词则匹配科目、章、节、要点与备注。
func (a *App) EntryVisible(e model.Entry) bool {
	if a.SubjectFilter != "" && e.Subject != a.SubjectFilter {
		return false
	}
	q := strings.TrimSpace(a.Query)
	if q == "" {
		return true
	}
	return EntryMatches(e, q)
}

// EntryMatches 报告一条记录里有没有 q（大小写不敏感）。
//
// 参数 q 是已经去掉首尾空白的搜索词。
func EntryMatches(e model.Entry, q string) bool {
	if model.QueryHit(e.Subject, q) || model.QueryHit(e.Chapter, q) ||
		model.QueryHit(e.Section, q) || model.QueryHit(e.Note, q) {
		return true
	}
	for _, p := range e.Points {
		if model.QueryHit(p.Text, q) {
			return true
		}
	}
	return false
}

// Subjects 汇总出现过的科目，用于输入框补全。
func (a *App) Subjects() []string {
	seen := make(map[string]bool, 8)
	out := make([]string, 0, 8)
	for _, d := range a.Days {
		for _, e := range d.Entries {
			if e.Subject == "" || seen[e.Subject] {
				continue
			}
			seen[e.Subject] = true
			out = append(out, e.Subject)
		}
	}
	return out
}

// ---------------------------------------------------------------- 编辑器

// OpenEditor 打开编辑器：entryID 为空表示新增。
//
// 正在按某个科目筛选时，新增的记录默认沿用这个科目。
func (a *App) OpenEditor(date, entryID string) {
	ed := &a.Editor
	ed.Open = true
	ed.Date = model.MustDate(date)
	if ed.Date.IsZero() {
		ed.Date, _ = model.ParseDate(a.Today)
	}
	ed.EntryID = entryID
	ed.Err = ""
	ed.Subject, ed.Chapter, ed.Section, ed.Points, ed.Note, ed.Minutes = "", "", "", "", "", 0

	if entryID == "" {
		ed.Subject = a.SubjectFilter
		return
	}
	day := a.Day(date)
	i := day.EntryByID(entryID)
	if i < 0 {
		return
	}
	e := day.Entries[i]
	ed.Subject = e.Subject
	ed.Chapter = e.Chapter
	ed.Section = e.Section
	ed.Note = e.Note
	ed.Minutes = float64(e.Minutes)
	lines := make([]string, 0, len(e.Points))
	for _, p := range e.Points {
		lines = append(lines, p.Text)
	}
	ed.Points = strings.Join(lines, "\n")
}

// SaveEditor 校验草稿并写入某一天。
//
// 校验失败时把原因写进 Editor.Err，弹窗不关闭。
func (a *App) SaveEditor() {
	// 每次保存都从「没有错误」开始，校验失败再写回原因。
	a.Editor.Err = ""
	subject := strings.TrimSpace(a.Editor.Subject)
	if subject == "" {
		a.Editor.Err = "请填写科目，例如「计算机组成原理」。"
		return
	}
	points := ParsePoints(a.Editor.Points)
	note := strings.TrimSpace(a.Editor.Note)
	if len(points) == 0 && note == "" {
		a.Editor.Err = "至少填写一条要点，或写一段备注。"
		return
	}

	date := a.Editor.Date.Format(model.DateLayout)
	di := a.DayIndex(date)
	if di < 0 {
		a.Upsert(model.Day{Date: date, Entries: []model.Entry{}})
		di = a.DayIndex(date)
		if di < 0 {
			a.Editor.Err = "找不到 " + date + " 这一天，保存失败。"
			return
		}
	}

	entry := model.Entry{
		ID:      a.Editor.EntryID,
		Subject: subject,
		Chapter: strings.TrimSpace(a.Editor.Chapter),
		Section: strings.TrimSpace(a.Editor.Section),
		Points:  points,
		Note:    note,
		Minutes: int(a.Editor.Minutes),
		At:      time.Now().Format("15:04"),
	}

	if entry.ID == "" {
		entry.ID = model.NewID()
		a.Days[di].Entries = append(a.Days[di].Entries, entry)
		a.Persist(date)
		a.Editor.Open = false
		a.Toast("已记录", subject+" · "+plural(len(points), "条要点"))
		return
	}

	j := a.Days[di].EntryByID(entry.ID)
	if j < 0 {
		a.Editor.Err = "这条记录已不存在，可能刚被删除。"
		return
	}
	old := a.Days[di].Entries[j]
	entry.At = old.At
	// 按文本对上旧要点，带过「已复习」与 ID，避免编辑一次就丢勾。
	for i := range entry.Points {
		for _, p := range old.Points {
			if p.Text == entry.Points[i].Text {
				entry.Points[i].ID = p.ID
				entry.Points[i].Done = p.Done
				break
			}
		}
	}
	a.Days[di].Entries[j] = entry
	a.Persist(date)
	a.Editor.Open = false
	a.Toast("已更新", subject+" · "+plural(len(entry.Points), "条要点"))
}

// ParsePoints 把多行文本拆成要点，顺手去掉 Markdown 的列表符号。
func ParsePoints(s string) []model.Point {
	var out []model.Point
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		for _, prefix := range []string{"- ", "* ", "• ", "· "} {
			line = strings.TrimPrefix(line, prefix)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, model.Point{ID: model.NewID(), Text: line})
	}
	return out
}

// ---------------------------------------------------------------- 删除

// AskDeleteEntry 弹出确认，删掉某一天里的某一条。
func (a *App) AskDeleteEntry(date, entryID string) {
	day := a.Day(date)
	i := day.EntryByID(entryID)
	if i < 0 {
		return
	}
	e := day.Entries[i]
	a.Confirm = ConfirmState{
		Open:    true,
		Title:   "删除这条记录？",
		Message: e.Subject + " · " + plural(len(e.Points), "条要点") + "，删除后无法恢复。",
		Yes:     "删除",
		YesFunc: func() { a.DeleteEntry(date, entryID) },
	}
}

// AskDeleteDay 弹出确认，删掉一整天的记录。待办不属于学记，不会一起删。
func (a *App) AskDeleteDay(date string) {
	day := a.Day(date)
	a.Confirm = ConfirmState{
		Open:    true,
		Title:   "删除 " + model.DisplayDate(date) + " 的全部记录？",
		Message: plural(len(day.Entries), "条学习记录") + "，删除后无法恢复。",
		Yes:     "删除",
		YesFunc: func() { a.DeleteDay(date) },
	}
}

// DeleteEntry 删除一条记录，并给一次撤销的机会。
func (a *App) DeleteEntry(date, entryID string) {
	di := a.DayIndex(date)
	if di < 0 {
		return
	}
	j := a.Days[di].EntryByID(entryID)
	if j < 0 {
		return
	}
	removed := a.Days[di].Entries[j]
	at := j
	a.Days[di].Entries = append(a.Days[di].Entries[:j], a.Days[di].Entries[j+1:]...)
	a.Persist(date)

	a.ToastUndo("已删除「"+removed.Subject+"」", "撤销", func() {
		k := a.DayIndex(date)
		if k < 0 {
			a.Upsert(model.Day{Date: date, Entries: []model.Entry{}})
			k = a.DayIndex(date)
			if k < 0 {
				return
			}
		}
		ents := a.Days[k].Entries
		if at > len(ents) {
			at = len(ents)
		}
		ents = append(ents, model.Entry{})
		copy(ents[at+1:], ents[at:])
		ents[at] = removed
		a.Days[k].Entries = ents
		a.Persist(date)
	})
}

// DeleteDay 清空一整天，并给一次撤销的机会。
func (a *App) DeleteDay(date string) {
	di := a.DayIndex(date)
	if di < 0 {
		return
	}
	backup := a.Days[di]
	a.Days[di].Entries = []model.Entry{}
	a.Persist(date)

	a.ToastUndo("已清空 "+model.DisplayDate(date), "撤销", func() {
		a.Upsert(backup)
		a.Persist(date)
	})
}

// plural 返回「3 条」这样的数量词。
func plural(n int, unit string) string {
	return fmt.Sprintf("%d %s", n, unit)
}
