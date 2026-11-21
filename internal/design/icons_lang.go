package design

import "github.com/egoist/mygo/ui"

// 语言选择用到的两个图标：中文用地球仪，英文用字母 A。
var (
	// IconGlobe 是地球仪：语言 / 中文。
	IconGlobe = ui.MustParseSVG([]byte(`<svg width="1em" height="1em" viewBox="0 0 24 24"><path d="M0 0h24v24H0z" fill="none" /><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><circle cx="12" cy="12" r="9" /><path d="M3 12h18M12 3c2.5 2.6 3.8 5.7 3.8 9S14.5 18.4 12 21c-2.5-2.6-3.8-5.7-3.8-9S9.5 5.6 12 3" /></g></svg>`))
	// IconLetterA 是字母 A：语言 / 英文。
	IconLetterA = ui.MustParseSVG([]byte(`<svg width="1em" height="1em" viewBox="0 0 24 24"><path d="M0 0h24v24H0z" fill="none" /><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path d="M6 20 12 4l6 16M9 14h6" /></g></svg>`))
)
