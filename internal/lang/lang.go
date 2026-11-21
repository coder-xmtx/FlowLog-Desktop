// Package lang 是界面文案：中文原文加一份英文词典。
//
// 界面代码里仍然直接写中文（它同时是词典的键），渲染前经过 T/Format 换成
// 当前语言。这样中文代码保持可读，英文只多一份映射；没有译文的句子退回中文。
//
// 语言本身由设置决定（Settings.Language），启动时应用到 MyGo（内置控件的
// 日历、日期框跟着切）和这里的 current。
package lang

import (
	"fmt"
	"strings"
)

// 支持的语言：中文是主语言，英文是译文。
const (
	// Chinese 是简体中文。
	Chinese = "zh-Hans"
	// English 是英文。
	English = "en"
)

// current 是当前界面语言；只在主线程写（启动与设置页切换时）。
var current = Chinese

// Current 返回当前界面语言。
func Current() string { return current }

// Set 切换界面语言；不认识的值回到中文。
func Set(tag string) {
	current = Normalize(tag)
}

// Normalize 把语言值收敛成 Chinese 或 English。
func Normalize(tag string) string {
	switch {
	case tag == English, strings.HasPrefix(tag, "en-"), strings.HasPrefix(tag, "en_"):
		return English
	default:
		return Chinese
	}
}

// IsEnglish 报告当前是不是英文。
func IsEnglish() bool { return current == English }

// T 返回中文原文在当前语言下的译文；当前是中文或没有译文时原样返回。
func T(zh string) string {
	if current == Chinese {
		return zh
	}
	if s, ok := en[zh]; ok {
		return s
	}
	return zh
}

// Format 先翻译模板、再按 Sprintf 填参数，用于带数字或日期的句子。
func Format(zh string, args ...any) string {
	return fmt.Sprintf(T(zh), args...)
}

// Has 报告英文词典里有没有这句（测试用）。
func Has(zh string) bool {
	_, ok := en[zh]
	return ok
}

// Missing 返回几句在英文词典里缺失的中文（测试里用来发现漏译）。
func Missing(zh ...string) []string {
	var out []string
	for _, s := range zh {
		if _, ok := en[s]; !ok {
			out = append(out, s)
		}
	}
	return out
}
