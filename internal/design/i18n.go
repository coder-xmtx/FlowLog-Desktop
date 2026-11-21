package design

import (
	"flowlog-desktop/internal/lang"

	"github.com/egoist/mygo/ui"
)

// 界面文案的多语言入口。
//
// 代码里照旧写中文原文，画到界面前经过 lang 换成当前语言；中文模式下
// 原样返回。标有 a11y 的标签（Label、Tooltip、Placeholder）也走这里，
// 这样读屏软件与占位提示同样会切换语言。

// Text 是 ui.Text 的多语言版本。
func Text(c *ui.Context, s string) ui.Element { return ui.Text(c, lang.T(s)) }

// Textf 先把格式串翻译好，再按参数格式化。
func Textf(c *ui.Context, format string, args ...any) ui.Element {
	return ui.Text(c, lang.Format(format, args...))
}

// T 返回一段文字在当前语言下的译文，用于 Placeholder、Tooltip、Label 等。
func T(s string) string { return lang.T(s) }

// Tf 返回带参数的译文。
func Tf(format string, args ...any) string { return lang.Format(format, args...) }
