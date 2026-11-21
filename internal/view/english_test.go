package view

import (
	"testing"
	"time"
	"unicode"

	"flowlog-desktop/internal/lang"

	"github.com/egoist/mygo/ui"
)

// TestEnglishPreviews 切到英文渲染每个页面，打印仍然含中文的文字。
//
// 它不判定失败：样例数据本身是中文（待办内容、随笔正文、学记科目），
// 需要人眼区分「漏翻的界面文案」与「用户数据」。
func TestEnglishPreviews(t *testing.T) {
	lang.Set(lang.English)
	defer lang.Set(lang.Chinese)

	a := sampleApp(t)
	for _, path := range []string{"/", "/study", "/notes", "/todo", "/settings"} {
		a.Router.Push(path)
		a.Router.Replace(path)
		tt := ui.NewTester(func(c *ui.Context) { Root(c, a) }, 1200, 1400)
		tt.Advance(16 * time.Millisecond)
		for _, text := range tt.Texts() {
			if hasHan(text) {
				t.Logf("%s: %s", path, text)
			}
		}
	}
}

// hasHan 报告字符串里有没有汉字。
func hasHan(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

// TestLangProbe 确认词典能查到几句常见文案。
func TestLangProbe(t *testing.T) {
	lang.Set(lang.English)
	defer lang.Set(lang.Chinese)
	for _, s := range []string{"在学记里查看全部", "全部待办", "晚上好", "首页", "写今天的学记"} {
		t.Logf("%s => %q", s, lang.T(s))
	}
}
