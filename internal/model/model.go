// Package model 定义流记的数据结构。
//
// 设计要点：
//   - 一天的记录是一个 Day，包含若干「学习条目」（Entry）与若干「待办」（Todo）。
//   - 每条学习条目是一个科目的一个主题，下面挂若干「要点」（Point），
//     对应以前 Markdown 里的一级标题、章节路径与列表项。
//   - 结构体字段全部带 json tag，方便直接落盘成 data/2026-11-13.json。
//   - 保持对早期版本（只有 diary 字符串 + todos）的兼容：LegacyDiary 字段只读，
//     由 store 在读取时迁移成 Entry。
package model

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// DateLayout 是磁盘上日期文件名与 Day.Date 的格式。
const DateLayout = "2006-01-02"

// Day 是一天的全部数据，对应 data/<日期>.json。
type Day struct {
	Date    string  `json:"date"`
	Entries []Entry `json:"entries,omitempty"`
	Todos   []Todo  `json:"todos,omitempty"`

	// LegacyDiary 是 v0.1 版本里那段裸 Markdown/纯文本日记。
	// 现代版本不再写入它，只在读取时把它迁移成一条 Entry。
	LegacyDiary string `json:"diary,omitempty"`
}

// Entry 是一条学习记录：某科目下的某个主题，以及这个主题下记住了什么。
type Entry struct {
	// ID 在一天之内唯一，用于编辑时定位与列表 key。
	ID string `json:"id"`
	// Subject 是科目/课程名，例如「计算机组成原理」。
	Subject string `json:"subject"`
	// Topic 是章节路径，用 " >> " 分隔，例如
	// 「第 4 章 存储系统 >> 第 8 节 虚拟存储器」。
	Topic string `json:"topic,omitempty"`
	// Points 是这一条下面的要点，等价于原来 Markdown 里的列表项。
	Points []Point `json:"points,omitempty"`
	// Note 是补充说明，自由文本。
	Note string `json:"note,omitempty"`
	// Minutes 是这次学习的时长（分钟），0 表示没记。
	Minutes int `json:"minutes,omitempty"`
	// At 是这条记录写入当刻的时间 "15:04"，用于排序与展示。
	At string `json:"at,omitempty"`
}

// Point 是一条学习要点。
type Point struct {
	ID   string `json:"id,omitempty"`
	Text string `json:"text"`
	// Done 表示这条要点已经复习/掌握，日记里可以打勾。
	Done bool `json:"done,omitempty"`
}

// Todo 是一条待办。
type Todo struct {
	ID   string `json:"id"`
	Text string `json:"text"`
	Done bool   `json:"done"`
	// At 是创建时间 "15:04"。
	At string `json:"at,omitempty"`
}

// Settings 是应用级别的偏好，存在 data/settings.json。
type Settings struct {
	// Theme 是外观："system"、"light" 或 "dark"。
	Theme string `json:"theme,omitempty"`
	// Accent 是强调色的十六进制值，空字符串表示用流记的品牌橙。
	Accent string `json:"accent,omitempty"`
}

// ThemeSystem / ThemeLight / ThemeDark 是 Settings.Theme 的三个取值。
const (
	ThemeSystem = "system"
	ThemeLight  = "light"
	ThemeDark   = "dark"
)

// NewID 返回一个 16 位十六进制随机 ID，用于 Entry / Point / Todo。
//
// 用随机数而不是自增或时间戳，是为了编辑、复制、合并数据文件时也不易冲突。
func NewID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand 在现代系统上不会失败；真失败了也要退化成可用值。
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

// Weekday 返回中文星期，例如「星期四」。
func Weekday(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return [...]string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}[int(t.Weekday())]
}

// WeekdayShort 返回「周日」「周一」这样的短星期。
func WeekdayShort(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return "周" + []string{"日", "一", "二", "三", "四", "五", "六"}[int(t.Weekday())]
}

// DisplayDate 返回「11 月 13 日」这样的展示用日期。
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

// DayNumber 返回日期是这一年的第几天，用来挑选每日寄语。
func DayNumber(t time.Time) int {
	return t.YearDay()
}

// HasContent 报告这一天有没有任何记录（学习条目或待办）。
func (d Day) HasContent() bool {
	return len(d.Entries) > 0 || len(d.Todos) > 0
}

// PointCount 返回这一天所有学习要点的条数。
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

// Minutes 返回这一天记录的学习总时长。
func (d Day) Minutes() int {
	n := 0
	for _, e := range d.Entries {
		n += e.Minutes
	}
	return n
}

// Subjects 返回这一天出现过的科目，按首次出现顺序去重。
func (d Day) Subjects() []string {
	seen := make(map[string]bool, len(d.Entries))
	out := make([]string, 0, len(d.Entries))
	for _, e := range d.Entries {
		if e.Subject == "" || seen[e.Subject] {
			continue
		}
		seen[e.Subject] = true
		out = append(out, e.Subject)
	}
	return out
}

// TopicParts 把 "第 4 章 存储系统 >> 第 8 节 虚拟存储器" 拆成面包屑。
//
// 支持 ">>" 与 "›" 两种分隔写法，方便从旧的 Markdown 直接粘进来；
// 不把 "/" 当分隔符，以免拆坏 "TCP/IP" 这类章节名。
func TopicParts(topic string) []string {
	replacer := strings.NewReplacer(">>", "\x00", "›", "\x00")
	parts := strings.Split(replacer.Replace(topic), "\x00")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// EntryByID 在一天里按 ID 找条目，返回下标，找不到为 -1。
func (d Day) EntryByID(id string) int {
	for i, e := range d.Entries {
		if e.ID == id {
			return i
		}
	}
	return -1
}

// Normalize 清理一天的数据，保证切片非 nil、日期规范。
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
