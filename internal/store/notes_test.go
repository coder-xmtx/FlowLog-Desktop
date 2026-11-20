package store

import (
	"testing"

	"flowlog-desktop/internal/model"
)

// TestNotesCRUD 检查随笔的读写与排序。
func TestNotesCRUD(t *testing.T) {
	s := newStore(t)
	notes := []model.Note{
		{ID: "n1", Date: "2026-11-01", Text: "第一条", At: "09:00"},
		{ID: "n2", Date: "2026-11-03", Text: "第三条", At: "21:00"},
		{ID: "n3", Date: "2026-11-03", Text: "同一天更晚", At: "22:00"},
	}
	for _, n := range notes {
		if err := s.SaveNote(n); err != nil {
			t.Fatalf("SaveNote(%s): %v", n.ID, err)
		}
	}
	got, err := s.Notes()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("读到 %d 条", len(got))
	}
	// 日期倒序，同一天按时刻倒序。
	want := []string{"n3", "n2", "n1"}
	for i, id := range want {
		if got[i].ID != id {
			t.Errorf("第 %d 条 = %q, 期望 %q", i, got[i].ID, id)
		}
	}

	n := got[0]
	n.Text = "改过的"
	if err := s.SaveNote(n); err != nil {
		t.Fatal(err)
	}
	got, _ = s.Notes()
	if len(got) != 3 || got[0].Text != "改过的" {
		t.Errorf("更新失败: %+v", got)
	}

	if err := s.DeleteNote("n2"); err != nil {
		t.Fatal(err)
	}
	got, _ = s.Notes()
	if len(got) != 2 {
		t.Errorf("删除后还剩 %d 条", len(got))
	}
}

// TestNotesInBackup 确认备份带上随笔。
func TestNotesInBackup(t *testing.T) {
	s := newStore(t)
	if err := s.SaveNote(model.Note{ID: "n1", Date: "2026-11-03", Text: "随笔内容", At: "21:00"}); err != nil {
		t.Fatal(err)
	}
	// 用临时文件走真实导出，再读回。
	path := t.TempDir() + "/backup.json"
	if err := s.ExportFile(path); err != nil {
		t.Fatal(err)
	}
	read, err := ReadBackup(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(read.Notes) != 1 || read.Notes[0].Text != "随笔内容" {
		t.Fatalf("备份里的随笔 = %+v", read.Notes)
	}
}
