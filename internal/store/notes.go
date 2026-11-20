package store

import (
	"context"
	"fmt"
	"strings"
	"time"

	"flowlog-desktop/internal/model"
)

// 随笔的读写：一张独立的表，按日期倒序排列，和学记、待办互不影响。

// Notes 读出全部随笔：按日期倒序，同一天按写入时间倒序。
func (s *Store) Notes() ([]model.Note, error) {
	rows, err := s.db.Query(context.Background(),
		`SELECT id, date, text, at FROM notes ORDER BY date DESC, created_at DESC, at DESC`)
	if err != nil {
		return nil, fmt.Errorf("读取随笔: %w", err)
	}
	out := make([]model.Note, 0, len(rows.Values))
	for _, row := range rows.Values {
		out = append(out, model.Note{
			ID:   text(row[0]),
			Date: text(row[1]),
			Text: text(row[2]),
			At:   text(row[3]),
		})
	}
	return out, nil
}

// SaveNote 写入或更新一条随笔。
func (s *Store) SaveNote(note model.Note) error {
	if note.ID == "" {
		return fmt.Errorf("随笔缺少 ID")
	}
	if err := validDate(note.Date); err != nil {
		return err
	}
	if strings.TrimSpace(note.Text) == "" {
		return fmt.Errorf("随笔内容为空")
	}
	_, err := s.db.Execute(context.Background(), `INSERT OR REPLACE INTO notes
		(id, date, text, at, created_at) VALUES (?, ?, ?, ?, ?)`,
		note.ID, note.Date, note.Text, note.At, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("保存随笔: %w", err)
	}
	return nil
}

// DeleteNote 删除一条随笔。
func (s *Store) DeleteNote(id string) error {
	if _, err := s.db.Execute(context.Background(), `DELETE FROM notes WHERE id = ?`, id); err != nil {
		return fmt.Errorf("删除随笔: %w", err)
	}
	return nil
}
