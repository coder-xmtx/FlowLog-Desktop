package lang

// 这些是为了英文语序重新组织的整句模板：中文仍然是原文，
// 拼接处换成 %s / %d，方便英文调整词序。
var extra = map[string]string{
	"你好，%s":            "Hello, %s",
	"已删除「%s」":          "Deleted \"%s\"",
	"删除 %s 的全部记录？":     "Delete every entry of %s?",
	"已清空 %s":           "Cleared %s",
	"已清除 %d 条已完成":      "Cleared %d completed todos",
	"找不到 %s 这一天，保存失败。": "No day %s to save to.",
	"「%s」，删除后无法恢复。":    "\"%s\", and this cannot be undone.",
	"随笔 · %s":          "Journal · %s",
	"%d 条要点":           "%d key points",
	"%d 条学习记录":         "%d study entries",
	"%s · %d 条要点":      "%s · %d key points",
	"%s · %d 条":        "%s · %d entries",
	"%s · %s":          "%s · %s",
	"本周末":              "This weekend",
	"已过期":              "Overdue",
	"已过期 · %s":         "Overdue · %s",
	"恢复完成":             "Restored",
	"已是最新版本":           "You're up to date",
	"当前版本 v%s":         "Current version v%s",
	"检查更新失败：%v":        "Update check failed: %v",
	"上次备份 %s · %s":     "Last backup %s · %s",
	"备份包含 %d 天记录与 %d 条待办。恢复会覆盖当前全部数据，且无法撤销。": "%d days of entries and %d todos are in the backup. Restoring replaces all current data and cannot be undone.",
	"恢复":   "Restore",
	"分钟":   " min",
	"小时":   " h",
	"语言":   "Language",
	"界面语言": "Interface language",
	"切换后界面立即生效，记录内容不会翻译": "Takes effect immediately. Your saved content is not translated.",
	"份":              "copies",
	"Bing 每日一图 · %s": "Bing daily image · %s",
	" · 已复习 %d":      " · %d reviewed",
	"发现新版本 v%s，安装后会自动重启":     "New version v%s is available and installs after a restart.",
	"从 %s 获取更新":              "Updates from %s",
	"MyGo %s · 原生 UI":        "MyGo %s · native UI",
	"Go %s · %s":             "Go %s · %s",
	"v%s 更新内容":               "What's new in v%s",
	"本次更新（v%s）":              "This update (v%s)",
	"写一条今天的记录":               "Write today's entry",
	"在学记里搜索":                 "Search the study log",
	"保存正在写的记录":               "Save the entry being written",
	"关闭弹窗":                   "Close dialogs",
	"首页 / 学记 / 随笔 / 待办 / 设置": "Home / Study Log / Journal / Todos / Settings",
	"条要点":                    "key points",
}

func init() {
	for k, v := range extra {
		if _, ok := en[k]; !ok {
			en[k] = v
		}
	}
}
