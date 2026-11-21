package design

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// TestNumberFieldLayout 量步进器三个控件的位置与大小，确认左减右加、同高。
func TestNumberFieldLayout(t *testing.T) {
	v := 45.0
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Row(c).Padding(20).Children(func() {
			NumberField(c, "nf", &v, 0, 600, 5, 72, nil)
		})
	}, 400, 120)
	tt.Advance(16 * time.Millisecond)

	for _, label := range []string{"减小", "nf", "增大"} {
		r, ok := tt.Find(label)
		if !ok {
			t.Fatalf("找不到 %q", label)
		}
		t.Logf("%s: x=%.1f y=%.1f w=%.1f h=%.1f", label, r.X, r.Y, r.W, r.H)
	}
}

// TestNumberFieldInField 重现「放在 Field 与固定宽度的列里」的场景。
func TestNumberFieldInField(t *testing.T) {
	v := 45.0
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Row(c).Padding(20).Children(func() {
			ui.Column(c).Width(168).Children(func() {
				ui.Field(c, "学习时长（分钟）", func() {
					NumberField(c, "nf", &v, 0, 600, 5, 56, nil)
				})
			})
		})
	}, 400, 160)
	tt.Advance(16 * time.Millisecond)

	for _, label := range []string{"减小", "nf", "增大"} {
		if r, ok := tt.Find(label); ok {
			t.Logf("%s: x=%.1f y=%.1f w=%.1f h=%.1f", label, r.X, r.Y, r.W, r.H)
		} else {
			t.Logf("%s: 找不到", label)
		}
	}
}
