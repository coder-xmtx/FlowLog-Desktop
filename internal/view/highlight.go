package view

import (
	"strings"
	"unicode"

	"github.com/egoist/mygo/ui"
)

// 这个文件负责「全文搜索命中哪些字」，以及把这些字在界面上标出来。
//
// 搜索是纯内存的：所有日记都在 app.days 里，匹配就是逐条字符串比较。
// 高亮用 MyGo 的富文本（RichText + Span 背景色），它会把整段当一个段落
// 排版，所以命中片段跨行、跨多个词都能正确绘制。

// textSegment 是一段文字，以及它是不是搜索命中的片段。
type textSegment struct {
	// Text 是这段原文。
	Text string
	// Match 为真表示这段是命中片段，界面上会加底色。
	Match bool
}

// splitHighlight 把 text 按 query 切成「普通 / 命中」交替的片段。
//
// 匹配大小写不敏感（BTree 能被 btree 搜到），按 rune 比较，
// 所以中文、emoji 都不会被切坏。query 为空或没有命中时，
// 返回一个 Match 为 false 的整段文本；命中不重叠，从左到右扫描。
//
// 参数 text 是原文，query 是用户输入的搜索词（内部会去掉首尾空白）。
func splitHighlight(text, query string) []textSegment {
	q := strings.TrimSpace(query)
	if q == "" || text == "" {
		return []textSegment{{Text: text}}
	}

	// 逐个 rune 转小写，保证下标与原文一一对应：
	// strings.ToLower 对个别字符会改变 rune 个数，那样就没法拿它切原文。
	runes := []rune(text)
	lowerText := make([]rune, len(runes))
	for i, r := range runes {
		lowerText[i] = unicode.ToLower(r)
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
		if !runesEqual(lowerText[i:i+len(lowerQuery)], lowerQuery) {
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

// queryHit 报告 text 里有没有 query（大小写不敏感）。
//
// 用来做「这一条要不要显示」的判断，比 splitHighlight 便宜。
func queryHit(text, query string) bool {
	q := strings.TrimSpace(query)
	if q == "" {
		return false
	}
	return strings.Contains(strings.ToLower(text), strings.ToLower(q))
}

// matchWeight 是命中片段的字重。
//
// 底色之外再压一点字重，深浅两套主题里都能一眼看到。
const matchWeight = 600

// highlightSpans 把 text 切成带底色的富文本片段。
//
// 参数 bg 是命中片段的底色；普通片段不设底色，沿用外层文字样式。
func highlightSpans(text, query string, bg ui.Color) []ui.Span {
	segments := splitHighlight(text, query)
	spans := make([]ui.Span, 0, len(segments))
	for _, seg := range segments {
		span := ui.Span{Text: seg.Text}
		if seg.Match {
			span.Background = bg
			span.Radius = 3
			span.Weight = matchWeight
		}
		spans = append(spans, span)
	}
	return spans
}

// highlightedText 构建一段可能带搜索高亮的文字。
//
// 没有搜索词、或这段文字没有命中时，返回普通 Text（更便宜，排版也与
// 其它文字完全一致）；命中时返回 RichText，命中的字加上底色。
// 参数 size 与 color 是文字的字号与颜色，返回的元素还可以继续链式
// 设置 SingleLine、Grow 等。
func highlightedText(c *ui.Context, text, query string, size float32, color ui.Color) ui.Element {
	t := c.Theme()
	segments := splitHighlight(text, query)
	matched := false
	for _, seg := range segments {
		if seg.Match {
			matched = true
			break
		}
	}
	if !matched {
		return ui.Text(c, text).FontSize(size).TextColor(color)
	}
	return ui.RichText(c, highlightSpans(text, query, highlightBG(t))...).
		FontSize(size).TextColor(color)
}
