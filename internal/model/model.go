// Package model 定义流记的数据结构。
//
// 一天的记录由若干「学习条目」（Entry）与若干「待办」（Todo）组成。
// 一条学习条目属于一个科目，按「章 / 节」两级定位，下面挂若干「要点」。
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

// Day 是一天的全部数据。
type Day struct {
	Date    string  `json:"date"`
	Entries []Entry `json:"entries"`
	Todos   []Todo  `json:"todos"`
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

// Todo 是一条待办。
type Todo struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
	At   string `json:"at,omitempty"`
}

// Settings 是应用级设置。
type Settings struct {
	// Theme 是 "light" 或 "dark"。
	Theme string `json:"theme"`
	// UserName 是用户在首次启动时留下的称呼。
	UserName string `json:"userName,omitempty"`
	// Onboarded 记录首次启动引导是否已经完成。
	Onboarded bool `json:"onboarded,omitempty"`
}

// ThemeLight / ThemeDark 是 Settings.Theme 的两个取值。
const (
	ThemeLight = "light"
	ThemeDark  = "dark"
)

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

// HasContent 报告这一天有没有任何记录。
func (d Day) HasContent() bool {
	return len(d.Entries) > 0 || len(d.Todos) > 0
}

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

// TodoProgress 返回已完成与总待办数。
func (d Day) TodoProgress() (done, total int) {
	for _, t := range d.Todos {
		if t.Done {
			done++
		}
	}
	return done, len(d.Todos)
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
// 老数据里可能把「章 >> 节」写在一个字段里，这里顺带兼容：
// 如果只有 Chapter 且包含 ">>"，就把它拆开显示。
func (e Entry) TopicParts() []string {
	chapter, section := strings.TrimSpace(e.Chapter), strings.TrimSpace(e.Section)
	if section == "" && strings.Contains(chapter, ">>") {
		parts := strings.Split(chapter, ">>")
		return trimAll(parts)
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
	if d.Todos == nil {
		d.Todos = []Todo{}
	}
	for i := range d.Entries {
		if d.Entries[i].Points == nil {
			d.Entries[i].Points = []Point{}
		}
	}
}
