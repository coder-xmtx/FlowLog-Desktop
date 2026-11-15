package view

import "github.com/egoist/mygo/ui"

// 图标全部是 24×24、2px 圆头描边的线条图标（Tabler Icons风格），
// 用 ui.MustParseSVG 在启动时解析一次，之后每帧复用同一个 *ui.SVG。
//
// ui.Icon 会用文字颜色重绘图标，所以颜色交给 .TextColor() 决定。
var (
	iconHome = ui.MustParseSVG([]byte(`
	<svg width="1em" height="1em" viewBox="0 0 24 24">
		<path d="M0 0h24v24H0z" fill="none" />
		<g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2">
			<path d="M5 12H3l9-9l9 9h-2M5 12v7a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2v-7" />
			<path d="M9 21v-6a2 2 0 0 1 2-2h2a2 2 0 0 1 2 2v6" />
		</g>
	</svg>
`))

	iconBook = ui.MustParseSVG([]byte(`
	<svg width="1em" height="1em" viewBox="0 0 24 24">
		<path d="M0 0h24v24H0z" fill="none" />
		<path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 4h11a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H6a1 1 0 0 1-1-1V5a1 1 0 0 1 1-1m3 0v18m4-14h2m-2 4h2" />
	</svg>
`))

	iconCheckCircle = ui.MustParseSVG([]byte(`
	<svg width="1em" height="1em" viewBox="0 0 24 24">
		<path d="M0 0h24v24H0z" fill="none" />
		<g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2">
			<path d="M3 12a9 9 0 1 0 18 0a9 9 0 1 0-18 0" />
			<path d="m9 12l2 2l4-4" />
		</g>
	</svg>
`))

	iconSettings = ui.MustParseSVG([]byte(`
	<svg width="1em" height="1em" viewBox="0 0 24 24">
		<path d="M0 0h24v24H0z" fill="none" />
		<g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2">
			<path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 0 0 2.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 0 0 1.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 0 0-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 0 0-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 0 0-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 0 0-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 0 0 1.066-2.573c-.94-1.543.826-3.31 2.37-2.37c1 .608 2.296.07 2.572-1.065" />
			<path d="M9 12a3 3 0 1 0 6 0a3 3 0 0 0-6 0" />
		</g>
	</svg>
`))

	iconPlus = ui.MustParseSVG([]byte(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 5v14"/><path d="M5 12h14"/></svg>`))

	iconPen = ui.MustParseSVG([]byte(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 20h9"/><path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7.5 18.5l-4 1 1-4z"/></svg>`))

	iconTrash = ui.MustParseSVG([]byte(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3.5 6h17"/><path d="M8.5 6V3.8h7V6"/><path d="M18.5 6 17.6 20H6.4L5.5 6"/><path d="M10 10.5v6"/><path d="M14 10.5v6"/></svg>`))

	iconSun = ui.MustParseSVG([]byte(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 17a5 5 0 1 0 0-10 5 5 0 0 0 0 10z"/><path d="M12 1.5v2"/><path d="M12 20.5v2"/><path d="M4.2 4.2 5.6 5.6"/><path d="M18.4 18.4l1.4 1.4"/><path d="M1.5 12h2"/><path d="M20.5 12h2"/><path d="M4.2 19.8 5.6 18.4"/><path d="M18.4 5.6l1.4-1.4"/></svg>`))

	iconMoon = ui.MustParseSVG([]byte(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z"/></svg>`))

	iconSearch = ui.MustParseSVG([]byte(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 19a8 8 0 1 0 0-16 8 8 0 0 0 0 16z"/><path d="M21 21l-4.3-4.3"/></svg>`))

	iconClose = ui.MustParseSVG([]byte(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="M6 6l12 12"/></svg>`))

	iconSparkle = ui.MustParseSVG([]byte(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3.5l1.7 4.3 4.3 1.7-4.3 1.7L12 15.5l-1.7-4.3L6 9.5l4.3-1.7z"/><path d="M18.5 16l.8 2 2 .8-2 .8-.8 2-.8-2-2-.8 2-.8z"/></svg>`))

	iconClock = ui.MustParseSVG([]byte(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 21.5a9.5 9.5 0 1 0 0-19 9.5 9.5 0 0 0 0 19z"/><path d="M12 7.5V12l3 1.8"/></svg>`))

	iconCopy = ui.MustParseSVG([]byte(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9.5 9.5h10v10h-10z"/><path d="M5 15V4.5h10"/></svg>`))

	iconCalendar = ui.MustParseSVG([]byte(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 6h16v14H4z"/><path d="M8 3v4"/><path d="M16 3v4"/><path d="M4 11h16"/></svg>`))

	iconInbox = ui.MustParseSVG([]byte(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3.5 12.5 6 4.5h12l2.5 8"/><path d="M3.5 12.5h5l1 2.5h5l1-2.5h5V19H3.5z"/></svg>`))

	iconDownload = ui.MustParseSVG([]byte(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3.5v12"/><path d="M7.5 11 12 15.5 16.5 11"/><path d="M4 19.5h16"/></svg>`))

	iconUpload = ui.MustParseSVG([]byte(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 15.5v-12"/><path d="M7.5 8 12 3.5 16.5 8"/><path d="M4 19.5h16"/></svg>`))

	iconAlert = ui.MustParseSVG([]byte(`<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10.3 3.9 2.6 17.2a2 2 0 0 0 1.7 3h15.4a2 2 0 0 0 1.7-3L13.7 3.9a2 2 0 0 0-3.4 0z"/><path d="M12 9v4.5"/><path d="M12 17h.01"/></svg>`))

	iconLeft = ui.MustParseSVG([]byte(`
	<svg width="1em" height="1em" viewBox="0 0 24 24">
		<path d="M0 0h24v24H0z" fill="none" />
		<path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m11 7l-5 5l5 5m6-10l-5 5l5 5" />
	</svg>
`))

	iconRight = ui.MustParseSVG([]byte(`
	<svg xmlns="http://www.w3.org/2000/svg" width="1em" height="1em" viewBox="0 0 24 24">
		<path d="M0 0h24v24H0z" fill="none" />
		<path fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m7 7l5 5l-5 5m6-10l5 5l-5 5" />
	</svg>
`))
)
