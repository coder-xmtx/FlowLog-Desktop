package pages

import (
	"testing"

	design "flowlog-desktop/internal/design"
)

// TestSplitHighlight 检查搜索片的切分：大小写、中文、重叠、没命中与空查询。
func TestSplitHighlight(t *testing.T) {
	cases := []struct {
		name  string
		text  string
		query string
		want  []textSegment
	}{
		{
			name:  "中文命中一次",
			text:  "虚拟存储器的地址映射",
			query: "存储器",
			want: []textSegment{
				{Text: "虚拟"},
				{Text: "存储器", Match: true},
				{Text: "的地址映射"},
			},
		},
		{
			name:  "大小写不敏感",
			text:  "BTree 和 btree 是同一棵树",
			query: "btree",
			want: []textSegment{
				{Text: "BTree", Match: true},
				{Text: " 和 "},
				{Text: "btree", Match: true},
				{Text: " 是同一棵树"},
			},
		},
		{
			name:  "开头与结尾都命中",
			text:  "页表页表",
			query: "页表",
			want: []textSegment{
				{Text: "页表", Match: true},
				{Text: "页表", Match: true},
			},
		},
		{
			name:  "没有命中时整段返回",
			text:  "数据结构",
			query: "操作系统",
			want:  []textSegment{{Text: "数据结构"}},
		},
		{
			name:  "空查询整段返回",
			text:  "数据结构",
			query: "   ",
			want:  []textSegment{{Text: "数据结构"}},
		},
		{
			name:  "查询比原文长",
			text:  "图",
			query: "图的基本概念",
			want:  []textSegment{{Text: "图"}},
		},
		{
			name:  "命中不重叠",
			text:  "aaaa",
			query: "aa",
			want: []textSegment{
				{Text: "aa", Match: true},
				{Text: "aa", Match: true},
			},
		},
	}

	for _, c := range cases {
		got := splitHighlight(c.text, c.query)
		if len(got) != len(c.want) {
			t.Errorf("%s: 切出 %d 段, 期望 %d 段: %+v", c.name, len(got), len(c.want), got)
			continue
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Errorf("%s: 第 %d 段 = %+v, 期望 %+v", c.name, i, got[i], c.want[i])
			}
		}
		// 拼回去必须还是原文：高亮不能改动内容。
		var joined string
		for _, seg := range got {
			joined += seg.Text
		}
		if joined != c.text {
			t.Errorf("%s: 拼回来 = %q, 原文 %q", c.name, joined, c.text)
		}
	}
}

// TestHighlightSpans 检查命中的片段带底色与文字色、普通片段不带。
func TestHighlightSpans(t *testing.T) {
	spans := highlightSpans("页表的结构", "页表", design.Light())
	if len(spans) != 2 {
		t.Fatalf("span 数 = %d: %+v", len(spans), spans)
	}
	if spans[0].Text != "页表" || spans[0].Background.A == 0 || spans[0].Color.A == 0 {
		t.Errorf("命中片段 = %+v", spans[0])
	}
	if spans[1].Text != "的结构" || spans[1].Background.A != 0 {
		t.Errorf("普通片段不该有底色: %+v", spans[1])
	}
	if spans[0].Weight != matchWeight {
		t.Errorf("命中片段字重 = %d", spans[0].Weight)
	}
}
