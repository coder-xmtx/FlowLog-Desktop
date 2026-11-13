package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"flowlog-desktop/internal/model"
)

// BackupFormat 是备份文件的格式版本。
const BackupFormat = 1

// Backup 是导出的备份文件内容：一个 JSON 文件带走全部数据。
type Backup struct {
	App        string         `json:"app"`
	Format     int            `json:"format"`
	ExportedAt time.Time      `json:"exportedAt"`
	Version    string         `json:"version,omitempty"`
	Settings   model.Settings `json:"settings"`
	Days       []model.Day    `json:"days"`
}

// Export 把全部数据写成一份 JSON 备份。
func (s *Store) Export(w io.Writer) error {
	days, err := s.Days()
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
	if b.Settings.Theme != model.ThemeDark {
		b.Settings.Theme = model.ThemeLight
	}
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
	return s.SaveSettings(b.Settings)
}

// legacyDay 是旧版 data/*.json 的形状。
//
// v0.1 只有 diary 字符串；v0.2 的章节写在 topic 里（"第 4 章 >> 第 8 节"）；
// v0.3 拆成 chapter / section。导入时三种都要认。
type legacyDay struct {
	Date    string        `json:"date"`
	Diary   string        `json:"diary"`
	Entries []legacyEntry `json:"entries"`
	Todos   []model.Todo  `json:"todos"`
}

type legacyEntry struct {
	model.Entry
	Topic string `json:"topic"`
}

// ImportJSONDir 读取旧版 data/*.json（一天一个文件）里的全部日记。
//
// 返回读到的日子与读不出来的文件名；用于从 JSON 版本一次性迁移。
func ImportJSONDir(dir string) (days []model.Day, bad []string, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, fmt.Errorf("读取目录: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		date := strings.TrimSuffix(e.Name(), ".json")
		if validDate(date) != nil {
			continue // settings.json 之类
		}
		raw, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			bad = append(bad, e.Name())
			continue
		}
		var legacy legacyDay
		if err := json.Unmarshal(raw, &legacy); err != nil {
			bad = append(bad, e.Name())
			continue
		}
		day := migrateLegacy(date, legacy)
		days = append(days, day)
	}
	SortDays(days)
	return days, bad, nil
}

// migrateLegacy 把旧版的一天换成当前的数据结构。
func migrateLegacy(date string, legacy legacyDay) model.Day {
	day := model.Day{Date: date, Todos: legacy.Todos}
	for _, le := range legacy.Entries {
		e := le.Entry
		if e.Chapter == "" && e.Section == "" {
			if chapter, section := splitTopic(le.Topic); chapter != "" {
				e.Chapter, e.Section = chapter, section
			}
		}
		day.Entries = append(day.Entries, e)
	}
	// v0.1：整段文字塞在 diary 里，转成一条「随手记」。
	if text := strings.TrimSpace(legacy.Diary); text != "" && len(day.Entries) == 0 {
		day.Entries = append(day.Entries, model.Entry{
			ID: model.NewID(), Subject: "随手记", Note: text, At: "00:00",
		})
	}
	day.Normalize()
	return day
}

// splitTopic 把 "第 4 章 存储系统 >> 第 8 节 虚拟存储器" 拆成章与节。
func splitTopic(topic string) (chapter, section string) {
	parts := strings.SplitN(topic, ">>", 2)
	chapter = strings.TrimSpace(parts[0])
	if len(parts) > 1 {
		section = strings.TrimSpace(parts[1])
	}
	return chapter, section
}

// ImportDays 把一批日子写进数据库（存在则覆盖）。
func (s *Store) ImportDays(days []model.Day) (int, error) {
	n := 0
	for _, day := range days {
		if err := validDate(day.Date); err != nil {
			continue
		}
		if err := s.SaveDay(day); err != nil {
			return n, err
		}
		n++
	}
	return n, nil
}

// IsNotBackup 报告错误是不是「文件不是备份」。
func IsNotBackup(err error) bool { return errors.Is(err, ErrNotBackup) }
