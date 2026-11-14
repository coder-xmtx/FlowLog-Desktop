// Package model 定义流记的数据结构。
//
// 两类数据互相独立：
//   - 日记：Day 里的一条条 Entry（科目 → 章 / 节 → 要点）。
//   - 待办：Todo，自带创建日期与可选的截止日期，不属于任何一天。
package model

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// DateLayout 是日期在主键、文件名与界面里的统一格式。
const DateLayout = "2006-01-02"

// Day 是一天的日记，只包含学习记录。
type Day struct {
	Date    string  `json:"date"`
	Entries []Entry `json:"entries"`
}

// Entry 是一条学习记录：某科目某个章节下记住的若干要点。
type Entry struct {
	ID      string  `json:"id"`
	Subject string  `json:"subject"`
	Chapter string  `json:"chapter,omitempty"`
	Section string  `json:"section,omitempty"`
	Points  []Point `json:"points"`
	Note    string  `json:"note,omitempty"`
	Minutes int     `json:"minutes,omitempty"`
	// At 是写入时刻的 "15:04"，用于同一天内的先后展示。
	At string `json:"at,omitempty"`
}

// Point 是一条学习要点。
type Point struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	// Done 表示这条要点已经复习。
	Done bool `json:"done,omitempty"`
}

// Todo 是一条待办，独立于日记。
type Todo struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
	// Created 是创建日期 "2026-11-14"。
	Created string `json:"created,omitempty"`
	// Due 是截止日期；空字符串表示没有截止日期。
	Due string `json:"due,omitempty"`
	// At 是创建时刻 "09:10"。
	At string `json:"at,omitempty"`
}

// DueState 描述一条待办相对今天的截止状态。
type DueState int

// 待办的截止状态，按紧急程度从高到低。
const (
	DueNone DueState = iota
	DueOverdue
	DueToday
	DueTomorrow
	DueLater
)

// DueStateOf 返回待办相对 today（"2006-01-02"）的截止状态。
func (t Todo) DueStateOf(today string) DueState {
	if strings.TrimSpace(t.Due) == "" {
		return DueNone
	}
	todayDate, err := ParseDate(today)
	if err != nil {
		return DueNone
	}
	due, err := ParseDate(t.Due)
	if err != nil {
		return DueNone
	}
	days := int(due.Sub(todayDate).Hours() / 24)
	switch {
	case days < 0:
		return DueOverdue
	case days == 0:
		return DueToday
	case days == 1:
		return DueTomorrow
	}
	return DueLater
}

// DueLabel 返回给人看的截止文案。
func (t Todo) DueLabel(today string) string {
	switch t.DueStateOf(today) {
	case DueOverdue:
		return "已过期 · " + DisplayDate(t.Due)
	case DueToday:
		return "今天到期"
	case DueTomorrow:
		return "明天到期"
	case DueLater:
		return DisplayDate(t.Due)
	}
	return ""
}

// Settings 是应用级设置。
type Settings struct {
	// Theme 是 "dark"、"light" 或 "system"。
	Theme string `json:"theme"`
	// UserName 是用户在首次启动时留下的称呼。
	UserName string `json:"userName,omitempty"`
	// Onboarded 记录首次启动引导是否已经完成。
	Onboarded bool `json:"onboarded,omitempty"`
}

// 外观取值。新装默认深色。
const (
	ThemeDark   = "dark"
	ThemeLight  = "light"
	ThemeSystem = "system"
)

// NormalizeTheme 把未知的外观值收敛成默认值（深色）。
func NormalizeTheme(theme string) string {
	switch theme {
	case ThemeLight, ThemeSystem, ThemeDark:
		return theme
	default:
		return ThemeDark
	}
}

// DefaultSettings 返回新装应用的默认设置。
func DefaultSettings() Settings {
	return Settings{Theme: ThemeDark}
}

// NewID 返回 16 位十六进制随机 ID。
//
// 用随机数而不是自增：数据在设备之间复制、合并时不会冲突。
func NewID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("t%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

// ParseDate 解析 "2006-01-02" 形式的日期。
func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation(DateLayout, s, time.Local)
}

// MustDate 返回 s 对应的日期，解析失败时返回零值。
func MustDate(s string) time.Time {
	t, err := ParseDate(s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// Today 返回今天的日期字符串。
func Today() string {
	return time.Now().Format(DateLayout)
}

// DayAfter 返回 date 之后 n 天的日期字符串。
func DayAfter(date string, n int) string {
	t := MustDate(date)
	if t.IsZero() {
		return date
	}
	return t.AddDate(0, 0, n).Format(DateLayout)
}

// Weekday 返回「星期四」。
func Weekday(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return [...]string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}[int(t.Weekday())]
}

// WeekdayShort 返回「周四」。
func WeekdayShort(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return "周" + []string{"日", "一", "二", "三", "四", "五", "六"}[int(t.Weekday())]
}

// DisplayDate 返回「11 月 13 日」。
func DisplayDate(s string) string {
	t := MustDate(s)
	if t.IsZero() {
		return s
	}
	return fmt.Sprintf("%d 月 %d 日", int(t.Month()), t.Day())
}

// DisplayDateFull 返回「2026 年 11 月 13 日」。
func DisplayDateFull(s string) string {
	t := MustDate(s)
	if t.IsZero() {
		return s
	}
	return fmt.Sprintf("%d 年 %d 月 %d 日", t.Year(), int(t.Month()), t.Day())
}

// RelativeDay 返回「今天」「昨天」「前天」或空字符串。
func RelativeDay(s string) string {
	t, err := ParseDate(s)
	if err != nil {
		return ""
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	diff := int(today.Sub(time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)).Hours() / 24)
	switch diff {
	case 0:
		return "今天"
	case 1:
		return "昨天"
	case 2:
		return "前天"
	}
	return ""
}

// DayNumber 返回日期是这一年的第几天。
func DayNumber(t time.Time) int { return t.YearDay() }

// HasEntries 报告这一天有没有学习记录。
func (d Day) HasEntries() bool { return len(d.Entries) > 0 }

// PointCount 返回所有学习要点的条数。
func (d Day) PointCount() int {
	n := 0
	for _, e := range d.Entries {
		n += len(e.Points)
	}
	return n
}

// DonePoints 返回已复习要点的条数。
func (d Day) DonePoints() int {
	n := 0
	for _, e := range d.Entries {
		for _, p := range e.Points {
			if p.Done {
				n++
			}
		}
	}
	return n
}

// EntryByID 按 ID 找条目，返回下标，找不到为 -1。
func (d Day) EntryByID(id string) int {
	for i, e := range d.Entries {
		if e.ID == id {
			return i
		}
	}
	return -1
}

// TopicParts 返回条目的定位路径：章、节，各自非空才出现。
//
// 老数据里可能把「章 >> 节」写在一个字段里，这里顺带兼容。
func (e Entry) TopicParts() []string {
	chapter, section := strings.TrimSpace(e.Chapter), strings.TrimSpace(e.Section)
	if section == "" && strings.Contains(chapter, ">>") {
		return trimAll(strings.Split(chapter, ">>"))
	}
	out := make([]string, 0, 2)
	if chapter != "" {
		out = append(out, chapter)
	}
	if section != "" {
		out = append(out, section)
	}
	return out
}

func trimAll(parts []string) []string {
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// Normalize 清空 nil 切片，保证序列化与遍历时不必到处判空。
func (d *Day) Normalize() {
	if d.Entries == nil {
		d.Entries = []Entry{}
	}
	for i := range d.Entries {
		if d.Entries[i].Points == nil {
			d.Entries[i].Points = []Point{}
		}
	}
}

// TodoProgress 返回一组待办的已完成数与总数。
func TodoProgress(todos []Todo) (done, total int) {
	for _, t := range todos {
		if t.Done {
			done++
		}
	}
	return done, len(todos)
}

// OpenTodos 返回还没完成的待办。
func OpenTodos(todos []Todo) []Todo {
	out := make([]Todo, 0, len(todos))
	for _, t := range todos {
		if !t.Done {
			out = append(out, t)
		}
	}
	return out
}

// DueSoon 返回需要今天关注的待办：已过期、今天到期、明天到期，只看未完成的。
//
// 顺序：过期的最先（越早越前），然后今天、明天；同一档按创建时间倒序。
func DueSoon(todos []Todo, today string) []Todo {
	out := make([]Todo, 0, 4)
	for _, t := range todos {
		switch t.DueStateOf(today) {
		case DueOverdue, DueToday, DueTomorrow:
			if !t.Done {
				out = append(out, t)
			}
		}
	}
	rank := func(t Todo) int {
		switch t.DueStateOf(today) {
		case DueOverdue:
			return 0
		case DueToday:
			return 1
		default:
			return 2
		}
	}
	// 稳定排序：同一档按 due 升序，再按创建时间倒序。
	for i := 1; i < len(out); i++ {
		for j := i; j > 0; j-- {
			a, b := out[j-1], out[j]
			ra, rb := rank(a), rank(b)
			swap := false
			switch {
			case ra != rb:
				swap = ra > rb
			case a.Due != b.Due:
				swap = a.Due > b.Due
			default:
				swap = a.Created < b.Created
			}
			if !swap {
				break
			}
			out[j-1], out[j] = b, a
		}
	}
	return out
}
