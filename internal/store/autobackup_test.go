package store

import (
	"path/filepath"
	"testing"
	"time"

	"flowlog-desktop/internal/model"
)

// TestAutoBackupWritesReadableBackup 确认自动备份写出来的是一份能读回去的备份。
func TestAutoBackupWritesReadableBackup(t *testing.T) {
	s := newStore(t)
	if err := s.SaveDay(sampleDay()); err != nil {
		t.Fatal(err)
	}
	if err := s.SaveTodo(sampleTodo()); err != nil {
		t.Fatal(err)
	}

	dir := filepath.Join(t.TempDir(), BackupDirName)
	now := time.Date(2026, 11, 18, 9, 30, 0, 0, time.Local)
	path, err := s.AutoBackup(dir, 14, now)
	if err != nil {
		t.Fatalf("AutoBackup: %v", err)
	}
	if filepath.Base(path) != "flowlog-auto-20261118-093000.json" {
		t.Errorf("备份文件名 = %q", filepath.Base(path))
	}

	b, err := ReadBackup(path)
	if err != nil {
		t.Fatalf("备份读不回来: %v", err)
	}
	if len(b.Days) != 1 || len(b.Days[0].Entries) != 1 {
		t.Errorf("备份里的日记 = %+v", b.Days)
	}
	if len(b.Todos) != 1 || b.Todos[0].ID != "t1" {
		t.Errorf("备份里的待办 = %+v", b.Todos)
	}
	if !b.Settings.AutoBackup || b.Settings.BackupKeep != model.DefaultBackupKeep {
		t.Errorf("备份里的设置 = %+v", b.Settings)
	}
}

// TestAutoBackupKeepsLatest 确认只保留最近 keep 份，且只清理自动备份。
func TestAutoBackupKeepsLatest(t *testing.T) {
	s := newStore(t)
	if err := s.SaveDay(sampleDay()); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), BackupDirName)
	base := time.Date(2026, 11, 1, 8, 0, 0, 0, time.Local)

	// 先放一份「恢复前」的备份，它不该被清理。
	if _, err := s.BackupBeforeRestore(dir, base); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		if _, err := s.AutoBackup(dir, 3, base.AddDate(0, 0, i)); err != nil {
			t.Fatalf("第 %d 次备份: %v", i, err)
		}
	}

	infos := listBackups(dir)
	auto, pre := 0, 0
	for _, info := range infos {
		switch {
		case filepath.Base(info.Path)[:len(autoBackupPrefix)] == autoBackupPrefix:
			auto++
		default:
			pre++
		}
	}
	if auto != 3 {
		t.Errorf("自动备份剩 %d 份, 期望 3", auto)
	}
	if pre != 1 {
		t.Errorf("恢复前的备份被误删了，剩 %d 份", pre)
	}

	latest, ok := LatestBackup(dir)
	if !ok {
		t.Fatal("找不到最新备份")
	}
	if latest.Name != "flowlog-auto-20261105-080000.json" {
		t.Errorf("最新备份 = %q", latest.Name)
	}
}

// TestBackupDue 确认「今天备份过没有」的判断。
func TestBackupDue(t *testing.T) {
	s := newStore(t)
	dir := filepath.Join(t.TempDir(), BackupDirName)
	now := time.Date(2026, 11, 18, 9, 30, 0, 0, time.Local)

	if !BackupDue(dir, now) {
		t.Error("目录还不存在时应该认为需要备份")
	}
	if _, err := s.AutoBackup(dir, 7, now); err != nil {
		t.Fatal(err)
	}
	if BackupDue(dir, now) {
		t.Error("今天已经备份过，不该再备份")
	}
	if !BackupDue(dir, now.AddDate(0, 0, 1)) {
		t.Error("到了第二天应该需要备份")
	}
}

// TestSettingsBackupDefaults 确认自动备份设置：默认开启、保留 14 份，且能被收敛。
func TestSettingsBackupDefaults(t *testing.T) {
	s := newStore(t)
	st, err := s.Settings()
	if err != nil {
		t.Fatal(err)
	}
	if !st.AutoBackup || st.BackupKeep != model.DefaultBackupKeep {
		t.Fatalf("默认设置 = %+v", st)
	}

	if err := s.SaveSettings(model.Settings{Theme: model.ThemeDark, AutoBackup: false, BackupKeep: 7}); err != nil {
		t.Fatal(err)
	}
	st, _ = s.Settings()
	if st.AutoBackup || st.BackupKeep != 7 {
		t.Fatalf("保存后 = %+v", st)
	}

	// 越界的份数收敛到合法区间。
	if err := s.SaveSettings(model.Settings{Theme: model.ThemeDark, BackupKeep: 0}); err != nil {
		t.Fatal(err)
	}
	st, _ = s.Settings()
	if st.BackupKeep != model.DefaultBackupKeep {
		t.Errorf("0 份被收敛成 %d, 期望默认值 %d", st.BackupKeep, model.DefaultBackupKeep)
	}
	if err := s.SaveSettings(model.Settings{Theme: model.ThemeDark, BackupKeep: 9999}); err != nil {
		t.Fatal(err)
	}
	st, _ = s.Settings()
	if st.BackupKeep != model.BackupKeepMax {
		t.Errorf("9999 份被收敛成 %d, 期望上限 %d", st.BackupKeep, model.BackupKeepMax)
	}
}
