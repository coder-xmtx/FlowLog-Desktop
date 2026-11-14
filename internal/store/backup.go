package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"flowlog-desktop/internal/model"
)

// BackupFormat 是备份文件的格式版本。
//
//	1  v0.3：待办放在 Day 里
//	2  v0.4：待办独立，带截止日期
const BackupFormat = 2

// Backup 是导出的备份文件内容：一个 JSON 文件带走全部数据。
type Backup struct {
	App        string         `json:"app"`
	Format     int            `json:"format"`
	ExportedAt time.Time      `json:"exportedAt"`
	Version    string         `json:"version,omitempty"`
	Settings   model.Settings `json:"settings"`
	Days       []model.Day    `json:"days"`
	Todos      []model.Todo   `json:"todos"`
}

// Export 把全部数据写成一份 JSON 备份。
func (s *Store) Export(w io.Writer) error {
	days, err := s.Days()
	if err != nil {
		return err
	}
	todos, err := s.Todos()
	if err != nil {
		return err
	}
	st, err := s.Settings()
	if err != nil {
		return err
	}
	b := Backup{
		App:        "FlowLog",
		Format:     BackupFormat,
		ExportedAt: time.Now(),
		Settings:   st,
		Days:       days,
		Todos:      todos,
	}
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(b); err != nil {
		return fmt.Errorf("写备份: %w", err)
	}
	return nil
}

// ExportFile 把备份写到 path。
func (s *Store) ExportFile(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("创建备份文件: %w", err)
	}
	defer f.Close()
	if err := s.Export(f); err != nil {
		return err
	}
	return f.Sync()
}

// ReadBackup 解析并校验一份备份文件。
func ReadBackup(path string) (*Backup, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取备份: %w", err)
	}
	var b Backup
	if err := json.Unmarshal(raw, &b); err != nil {
		return nil, ErrNotBackup
	}
	if b.App != "FlowLog" || b.Format < 1 {
		return nil, ErrNotBackup
	}
	for i := range b.Days {
		if err := validDate(b.Days[i].Date); err != nil {
			return nil, fmt.Errorf("备份里有一天的日期不合法: %w", err)
		}
		b.Days[i].Normalize()
	}
	for i := range b.Todos {
		if b.Todos[i].Due != "" {
			if err := validDate(b.Todos[i].Due); err != nil {
				return nil, fmt.Errorf("备份里有一条待办的截止日期不合法: %w", err)
			}
		}
		if b.Todos[i].ID == "" {
			b.Todos[i].ID = model.NewID()
		}
	}
	b.Settings.Theme = model.NormalizeTheme(b.Settings.Theme)
	return &b, nil
}

// Restore 用备份替换当前全部数据。
func (s *Store) Restore(b *Backup) error {
	if b == nil {
		return ErrNotBackup
	}
	if err := s.DeleteAll(); err != nil {
		return err
	}
	for _, day := range b.Days {
		if err := s.SaveDay(day); err != nil {
			return err
		}
	}
	for _, todo := range b.Todos {
		if err := s.SaveTodo(todo); err != nil {
			return err
		}
	}
	return s.SaveSettings(b.Settings)
}

// IsNotBackup 报告错误是不是「文件不是备份」。
func IsNotBackup(err error) bool { return errors.Is(err, ErrNotBackup) }
