// 自动备份：每次启动（或用户手动点一下）把全部数据导出成一份 JSON，
// 放在数据目录的 backups/ 下，只保留最近若干份。
//
// 备份是 JSON 而不是数据库文件：用户可以直接看、直接改，出问题时用
// 设置页的「从备份恢复」读回去，和手动导出的备份是同一个格式。
package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// BackupDirName 是数据目录下存放自动备份的子目录名。
const BackupDirName = "backups"

// 自动备份的文件名前缀。
const (
	// autoBackupPrefix 是每天一次自动备份的前缀。
	autoBackupPrefix = "flowlog-auto-"
	// preRestorePrefix 是「从备份恢复之前先存一份」的前缀。
	preRestorePrefix = "flowlog-before-restore-"
	// backupExt 是备份文件的扩展名。
	backupExt = ".json"
)

// backupStampLayout 是文件名里的时间戳格式：20261118-150405。
const backupStampLayout = "20060102-150405"

// BackupInfo 描述磁盘上的一份备份文件。
type BackupInfo struct {
	// Path 是文件的完整路径。
	Path string
	// Name 是文件名。
	Name string
	// Size 是文件大小（字节）。
	Size int64
	// ModTime 是文件的修改时间。
	ModTime time.Time
}

// AutoBackupDir 返回数据库所在目录下的备份目录。
//
// 参数 dbPath 是数据库文件路径；备份与数据库放在一起，
// 复制数据目录就能把日记和备份一起带走。
func AutoBackupDir(dbPath string) string {
	return filepath.Join(filepath.Dir(dbPath), BackupDirName)
}

// AutoBackup 导出一份自动备份，并按 keep 清理旧文件。
//
// 参数 dir 是备份目录（不存在会创建），keep 是保留份数（至少 1），
// now 决定文件名里的时间戳，便于测试。返回写出的文件路径。
func (s *Store) AutoBackup(dir string, keep int, now time.Time) (string, error) {
	if keep < 1 {
		keep = 1
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("创建备份目录: %w", err)
	}
	path := filepath.Join(dir, autoBackupPrefix+now.Format(backupStampLayout)+backupExt)
	// 同一秒内连续备份（测试或手快）时不要覆盖已有文件。
	for i := 2; fileExists(path); i++ {
		path = filepath.Join(dir, fmt.Sprintf("%s%s-%d%s",
			autoBackupPrefix, now.Format(backupStampLayout), i, backupExt))
	}
	if err := s.ExportFile(path); err != nil {
		return "", err
	}
	// 清理失败不算备份失败：文件已经写好了，最多是多留几份。
	_ = pruneBackups(dir, keep)
	return path, nil
}

// BackupBeforeRestore 在覆盖数据之前先存一份，名字能看出来是恢复前的状态。
//
// 参数 dir 是备份目录，now 决定时间戳；返回写出的文件路径。
func (s *Store) BackupBeforeRestore(dir string, now time.Time) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("创建备份目录: %w", err)
	}
	path := filepath.Join(dir, preRestorePrefix+now.Format(backupStampLayout)+backupExt)
	if err := s.ExportFile(path); err != nil {
		return "", err
	}
	return path, nil
}

// BackupDue 报告 dir 里有没有 now 这一天的自动备份。
//
// 每天只备份一次：启动时先问这一句，有就直接跳过。
func BackupDue(dir string, now time.Time) bool {
	prefix := autoBackupPrefix + now.Format("20060102")
	infos := listBackups(dir)
	for _, info := range infos {
		if strings.HasPrefix(info.Name, prefix) {
			return false
		}
	}
	return true
}

// LatestBackup 返回 dir 里最新的一份备份。
//
// 按文件名里的时间戳排序（不是修改时间，也不是字典序：两种前缀的
// 字典序和时间顺序并不一致），第二个返回值表示有没有。
func LatestBackup(dir string) (BackupInfo, bool) {
	infos := listBackups(dir)
	if len(infos) == 0 {
		return BackupInfo{}, false
	}
	latest := infos[0]
	for _, info := range infos[1:] {
		if backupStamp(info.Name).After(backupStamp(latest.Name)) {
			latest = info
		}
	}
	return latest, true
}

// backupStamp 从文件名里取出时间戳；取不到时返回零值。
//
// 文件名形如 flowlog-auto-20261118-093000.json 或
// flowlog-auto-20261118-093000-2.json（同一秒里重名时的去重后缀）。
func backupStamp(name string) time.Time {
	base := strings.TrimSuffix(name, backupExt)
	if n := len(base); n > len(backupStampLayout) && base[n-2] == '-' {
		base = base[:n-2]
	}
	if len(base) < len(backupStampLayout) {
		return time.Time{}
	}
	stamp := base[len(base)-len(backupStampLayout):]
	t, err := time.ParseInLocation(backupStampLayout, stamp, time.Local)
	if err != nil {
		return time.Time{}
	}
	return t
}

// listBackups 列出 dir 里所有的备份文件（两种前缀都算）。
//
// 目录不存在、读不了时返回 nil：调用方都按「没有备份」处理。
func listBackups(dir string) []BackupInfo {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	out := make([]BackupInfo, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, backupExt) {
			continue
		}
		if !strings.HasPrefix(name, autoBackupPrefix) && !strings.HasPrefix(name, preRestorePrefix) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		out = append(out, BackupInfo{
			Path:    filepath.Join(dir, name),
			Name:    name,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}
	return out
}

// pruneBackups 只保留最新的 keep 份自动备份。
//
// 只清理自动备份（flowlog-auto-*），恢复前的那份与用户手动导出的文件
// 一律不动——它们要么是安全网，要么是用户自己放的。
func pruneBackups(dir string, keep int) error {
	infos := listBackups(dir)
	auto := infos[:0]
	for _, info := range infos {
		if strings.HasPrefix(info.Name, autoBackupPrefix) {
			auto = append(auto, info)
		}
	}
	if len(auto) <= keep {
		return nil
	}
	sort.Slice(auto, func(i, j int) bool {
		return backupStamp(auto[i].Name).Before(backupStamp(auto[j].Name))
	})
	for _, info := range auto[:len(auto)-keep] {
		if err := os.Remove(info.Path); err != nil {
			return err
		}
	}
	return nil
}

// fileExists 报告路径上是否已经有文件。
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
