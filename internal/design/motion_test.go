package design

import (
	"image/color"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// TestHoverLayerAnimates 确认 hover 的 state layer 是逐帧过渡的，而不是一下子跳变。
//
// 做法：渲染一个方块，指针移进去后分别取 30ms 与 400ms 的像素颜色。
// 30ms 时应该还在中间色，400ms 时已经到达目标色。
func TestHoverLayerAnimates(t *testing.T) {
	p := Light()
	base := p.SurfaceContainerHigh
	hover := base.Mix(p.OnSurface, 0.18) // 明显一点的目标色，便于判定

	var box ui.Element
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Padding(20).Children(func() {
			box = ui.ButtonBase(c.Key("probe")).Size(200, 80).Radius(RadiusSM)
			bg := base
			if box.Hovered() {
				bg = hover
			}
			box.Background(box.AnimateColor("hover", bg, MotionState))
		})
	}, 260, 140)
	tt.Advance(16 * time.Millisecond)

	start := sample(tt, 100, 60)
	tt.Move(100, 60)
	tt.Advance(30 * time.Millisecond)
	mid := sample(tt, 100, 60)
	tt.Advance(400 * time.Millisecond)
	end := sample(tt, 100, 60)

	if start == end {
		t.Fatalf("悬停前后颜色一样，动画没生效: %v", start)
	}
	if mid == start || mid == end {
		t.Fatalf("30ms 时应该还在过渡中间: start=%v mid=%v end=%v", start, mid, end)
	}
	if !nearColor(end, rgbOf(hover), 6) {
		t.Errorf("400ms 后应到达目标色: end=%v 期望=%v", end, rgbOf(hover))
	}
}

// sample 取测试器画面上的一个像素。
func sample(tt *ui.Tester, x, y int) color.RGBA {
	return tt.Image().RGBAAt(x, y)
}

// rgbOf 把不透明的主题色转成 color.RGBA。
func rgbOf(c ui.Color) color.RGBA {
	return color.RGBA{R: c.R, G: c.G, B: c.B, A: 255}
}

// nearColor 判断两个颜色每个通道是否接近。
func nearColor(a, b color.RGBA, tol int) bool {
	diff := func(x, y uint8) int {
		if x > y {
			return int(x - y)
		}
		return int(y - x)
	}
	return diff(a.R, b.R) <= tol && diff(a.G, b.G) <= tol && diff(a.B, b.B) <= tol
}
