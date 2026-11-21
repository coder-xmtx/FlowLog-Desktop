package model

import (
	"fmt"
	"time"

	"flowlog-desktop/internal/lang"
)

// 日期与时间的显示：中文用「11 月 13 日」这种写法，英文用 "Nov 13"。

// enMonths 是英文月份缩写。
var enMonths = [...]string{
	"Jan", "Feb", "Mar", "Apr", "May", "Jun",
	"Jul", "Aug", "Sep", "Oct", "Nov", "Dec",
}

// enWeekdays 是英文星期的全称，按 time.Weekday 排列。
var enWeekdays = [...]string{
	"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday",
}

// DisplayDate 返回「11 月 13 日」，英文为 "Nov 13"。
func DisplayDate(s string) string {
	t := MustDate(s)
	if t.IsZero() {
		return s
	}
	if lang.IsEnglish() {
		return fmt.Sprintf("%s %d", enMonths[int(t.Month())-1], t.Day())
	}
	return fmt.Sprintf("%d 月 %d 日", int(t.Month()), t.Day())
}

// DisplayDateFull 返回「2026 年 11 月 13 日」，英文为 "Nov 13, 2026"。
func DisplayDateFull(s string) string {
	t := MustDate(s)
	if t.IsZero() {
		return s
	}
	if lang.IsEnglish() {
		return fmt.Sprintf("%s %d, %d", enMonths[int(t.Month())-1], t.Day(), t.Year())
	}
	return fmt.Sprintf("%d 年 %d 月 %d 日", t.Year(), int(t.Month()), t.Day())
}

// MonthTitle 返回月历的标题「2026 年 11 月」，英文为 "November 2026"。
func MonthTitle(year int, month time.Month) string {
	if lang.IsEnglish() {
		return enMonths[int(month)-1] + " " + fmt.Sprint(year)
	}
	return fmt.Sprintf("%d 年 %d 月", year, int(month))
}

// WeekdayHead 返回月历表头的星期缩写：「一」…「日」或 "Mo"…"Su"。
func WeekdayHead(t time.Time) string {
	if lang.IsEnglish() {
		return [...]string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}[int(t.Weekday())]
	}
	return [...]string{"日", "一", "二", "三", "四", "五", "六"}[int(t.Weekday())]
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
		return lang.T("今天")
	case 1:
		return lang.T("昨天")
	case 2:
		return lang.T("前天")
	}
	return ""
}

// DayLabel 是月历里一天的读屏文字：「11 月 13 日」或 "Nov 13"。
func DayLabel(year int, month time.Month, day int) string {
	return DisplayDate(time.Date(year, month, day, 0, 0, 0, 0, time.Local).Format(DateLayout))
}
