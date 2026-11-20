package pages

import (
	"strings"
	"unicode"

	design "flowlog-desktop/internal/design"

	"github.com/egoist/mygo/ui"
)

// 这个文件负责把搜索命中的字切出来并在界面上标出来。
//
// 匹配用 rune 比较、大小写不敏感；高亮用 MyGo 的富文本 span 背景色，
// 整段仍按一个段落排版，跨行也连续。

// matchWeight 是命中片段的字重，比正文重一档。
const matchWeight = 600

// textSegment 是一段文字，以及它是不是搜索命中的片段。
type textSegment struct {
	// Text 是这段原文。
	Text string
	// Match 为真表示这段是命中片段。
	Match bool
}

// splitHighlight 把 text 按 query 切成「普通 / 命中」交替的片段。
//
// 命中不重叠，从左到右扫描；query 为空或没有命中时返回整段普通文本。
// 参数 text 是原文，query 是搜索词（首尾空白会被去掉）。
func splitHighlight(text, query string) []textSegment {
	q := strings.TrimSpace(query)
	if q == "" || text == "" {
		return []textSegment{{Text: text}}
	}

	// 逐个 rune 转小写，保证下标与原文一一对应。
	runes := []rune(text)
	lower := make([]rune, len(runes))
	for i, r := range runes {
		lower[i] = unicode.ToLower(r)
	}
	queryRunes := []rune(q)
	lowerQuery := make([]rune, len(queryRunes))
	for i, r := range queryRunes {
		lowerQuery[i] = unicode.ToLower(r)
	}
	if len(lowerQuery) == 0 || len(lowerQuery) > len(runes) {
		return []textSegment{{Text: text}}
	}

	var out []textSegment
	last := 0
	for i := 0; i+len(lowerQuery) <= len(runes); {
		if !runesEqual(lower[i:i+len(lowerQuery)], lowerQuery) {
			i++
			continue
		}
		if last < i {
			out = append(out, textSegment{Text: string(runes[last:i])})
		}
		out = append(out, textSegment{Text: string(runes[i : i+len(lowerQuery)]), Match: true})
		i += len(lowerQuery)
		last = i
	}
	if last < len(runes) {
		out = append(out, textSegment{Text: string(runes[last:])})
	}
	if len(out) == 0 {
		return []textSegment{{Text: text}}
	}
	return out
}

// runesEqual 比较两段 rune 是否相同。
func runesEqual(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// highlightSpans 把 text 切成带底色的富文本片段。
//
// 参数 p 提供高亮色与文字色；普通片段不设样式，沿用外层文字样式。
func highlightSpans(text, query string, p design.Palette) []ui.Span {
	segments := splitHighlight(text, query)
	spans := make([]ui.Span, 0, len(segments))
	for _, seg := range segments {
		span := ui.Span{Text: seg.Text}
		if seg.Match {
			span.Background = p.Highlight
			span.Color = p.OnHighlight
			span.Radius = 3
			span.Weight = matchWeight
		}
		spans = append(spans, span)
	}
	return spans
}

// highlightedText 构建一段可能带搜索高亮的文字。
//
// 没有搜索词、或这段文字没有命中时返回普通 Text（更便宜，排版也完全一致）；
// 命中时返回 RichText，命中的字加琥珀底色。参数 size/color 是字号与颜色。
func highlightedText(c *ui.Context, text, query string, size float32, color ui.Color) ui.Element {
	p := design.PaletteOf(c)
	matched := false
	for _, seg := range splitHighlight(text, query) {
		if seg.Match {
			matched = true
			break
		}
	}
	if !matched {
		return ui.Text(c, text).FontSize(size).TextColor(color)
	}
	return ui.RichText(c, highlightSpans(text, query, p)...).
		FontSize(size).TextColor(color)
}
