package model

import "flowlog-desktop/internal/lang"

// ShortDateTimeLayout 是「1 月 2 日 15:04」的布局，英文为 "Jan 2 15:04"。
func ShortDateTimeLayout() string {
	if lang.IsEnglish() {
		return "Jan 2 15:04"
	}
	return "1 月 2 日 15:04"
}

// ShortDateLayout 是「1 月 2 日」的布局，英文为 "Jan 2"。
func ShortDateLayout() string {
	if lang.IsEnglish() {
		return "Jan 2"
	}
	return "1月2日"
}
