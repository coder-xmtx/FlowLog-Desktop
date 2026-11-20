package design

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// 动效时长参照 Material 3 的 motion tokens：状态变化 150ms，按下 90ms，
// 页面切换 220ms。全部走 ease-out，进入时快、停下时慢，观感更轻。
var (
	// MotionState 是 hover / 聚焦 / 选中这类「原地」状态变化。
	MotionState = ui.Animation{Duration: 150 * time.Millisecond, Ease: ui.EaseOut}
	// MotionPress 是按下的反馈，比 hover 更快。
	MotionPress = ui.Animation{Duration: 90 * time.Millisecond, Ease: ui.EaseOut}
	// MotionSwitch 是页面、弹层切换。
	MotionSwitch = ui.Animation{Duration: 220 * time.Millisecond, Ease: ui.EaseInOut}
)

// HoverSurface 返回一个随鼠标进出平滑过渡的背景色。
//
// 参数 el 是参与动画的元素（通常是这一行/这张卡本身），key 区分同一个元素上的
// 多组动画，base 是常态底色，hover 是悬停底色。元素没被悬停时向 base 过渡，
// 悬停时向 hover 过渡，中间由 MyGo 逐帧插值，所以不会「啪」地跳变。
//
// MyGo 在系统开启「减少动态效果」时会自动跳到目标值，不需要额外判断。
func HoverSurface(el ui.Element, key any, base, hover ui.Color) ui.Color {
	if el.Hovered() {
		return el.AnimateColor(key, hover, MotionState)
	}
	return el.AnimateColor(key, base, MotionState)
}

// HoverLayer 是 MD3 的 state layer：在底色上叠一层低透明度的强调色。
//
// 参数 tint 通常传主题的 Primary 或 OnSurface；hovered 决定这层是否浮现。
// 返回的颜色带 alpha，直接当背景色用即可（父容器的底色会透出来）。
func HoverLayer(el ui.Element, key any, tint ui.Color, hovered bool) ui.Color {
	target := tint.Alpha(0)
	if hovered {
		target = tint.Alpha(0.08)
	}
	return el.AnimateColor(key, target, MotionState)
}

// PressLayer 是按下时的 state layer，比 hover 深一点。
func PressLayer(el ui.Element, key any, tint ui.Color, pressed bool) ui.Color {
	target := tint.Alpha(0)
	if pressed {
		target = tint.Alpha(0.12)
	}
	return el.AnimateColor(key, target, MotionPress)
}

// Fade 把颜色按 t（0–1）插值，用于自己算的过渡。
func Fade(from, to ui.Color, t float32) ui.Color {
	return from.Mix(to, t)
}
