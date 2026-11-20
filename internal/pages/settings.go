package pages

import (
	"fmt"
	"strings"

	"flowlog-desktop/internal/app"
	design "flowlog-desktop/internal/design"
	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// Settings 是设置页：外观、账户与数据；关于本应用的信息在最后
func Settings(c *ui.Context, a *app.App) {
	p := design.PaletteOf(c)
	days, entries, points, todos := a.Summary()

	ui.Column(c).Grow(1).MinWidth(0).Children(func() {
		design.PageHeader(c, "设置", "外观、账户与数据；关于本应用的信息在最后", nil)

		ui.Scroll(c).Grow(1).Padding(6, 32, 28).Gap(16).Children(func() {
			settingsAccount(c, a)
			settingsAppearance(c, a)
			settingsData(c, a, days, entries, points, todos)
			settingsShortcuts(c)
			settingsAbout(c, a)
		})
		_ = p
	})
}

// settingsAccount 是称呼。
func settingsAccount(c *ui.Context, a *app.App) {
	p := design.PaletteOf(c)
	design.SectionCard(c, "账户", "", func() {
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			in := ui.TextInput(c.Key("settings-name"), &a.Settings.UserName).
				Label("称呼").MaxLength(24).Placeholder("你的称呼").Height(design.ControlHeight).Grow(1)
			if in.Submitted() {
				a.SaveSettings()
				a.Toast("已保存", "称呼已更新")
			}
			design.FilledButton(c, "save-name", "保存", nil, func() {
				a.SaveSettings()
				a.Toast("已保存", "称呼已更新")
			})
		})
		ui.Text(c, "首页的问候会用到这个称呼，随时可以修改").
			FontSize(design.FontCaption).TextColor(p.OnSurfaceVariant)
	})
}

// settingsAppearance 是深色 / 浅色 / 跟随系统。
func settingsAppearance(c *ui.Context, a *app.App) {
	p := design.PaletteOf(c)
	design.SectionCard(c, "外观", "", func() {
		// 分段控件绑定下标；Changed 报告新选择，此时 choice 已是新值。
		choice := 0
		switch a.Settings.Theme {
		case model.ThemeLight:
			choice = 1
		case model.ThemeSystem:
			choice = 2
		}
		seg := design.IconSegmented(c, "settings-theme", &choice, 400,
			design.SegmentOption{Icon: design.IconMoon, Label: "深色"},
			design.SegmentOption{Icon: design.IconSun, Label: "浅色"},
			design.SegmentOption{Icon: design.IconMonitor, Label: "跟随系统"},
		)
		if seg.Changed() {
			theme := model.ThemeDark
			switch choice {
			case 1:
				theme = model.ThemeLight
			case 2:
				theme = model.ThemeSystem
			}
			a.SetTheme(theme)
		}
		ui.Text(c, "默认深色；选择「跟随系统」后，应用会随系统的深浅色切换").
			FontSize(design.FontCaption).TextColor(p.OnSurfaceVariant)
	})
}

// settingsData 是数据库信息、备份与自动备份。
//
// 参数 days、entries、points、todos 是全库的规模统计。
func settingsData(c *ui.Context, a *app.App, days, entries, points, todos int) {
	p := design.PaletteOf(c)
	path := a.Store.Path()

	design.SectionCard(c, "数据", "本地 SQLite 数据库", func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Textf(c, "%d 天 · %d 条记录 · %d 条要点 · %d 条待办",
				days, entries, points, todos).FontSize(design.FontBody).TextColor(p.OnSurface)
			ui.Spacer(c)
			ui.Text(c, app.FormatSize(a.Store.Size())).FontSize(design.FontCaption).
				TextColor(p.OnSurfaceVariant)
		})

		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Grow(1).Padding(9, 12).Radius(design.RadiusSM).
				Background(p.SurfaceContainer).Children(func() {
				ui.Text(c, path).FontSize(design.FontCaption).Font("monospace").
					Selectable().SingleLine().Tooltip(path).TextColor(p.OnSurfaceVariant)
			})
			design.OutlinedButton(c, "copy-db-path", "复制路径", design.IconCopy, func() {
				c.WriteClipboard(path)
				a.Toast("已复制路径", path)
			})
		})

		ui.Row(c).Gap(8).Wrap().Children(func() {
			design.OutlinedButton(c, "export-backup", "导出备份", design.IconDownload, a.ExportBackup)
			design.OutlinedButton(c, "restore-backup", "从备份恢复", design.IconUpload, a.ChooseBackup)
		})

		ui.Text(c, "备份是一个 JSON 文件，包含全部记录、待办与设置，可以复制到别的机器上恢复").
			FontSize(design.FontCaption).TextColor(p.OnSurfaceVariant).LineHeight(1.6)

		autoBackupRow(c, a)
	})
}

// autoBackupRow 是「自动备份」那一块：开关、保留份数、最近一次备份与手动备份。
func autoBackupRow(c *ui.Context, a *app.App) {
	p := design.PaletteOf(c)
	design.Divider(c)

	ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
		ui.Text(c, "自动备份").FontSize(design.FontBody).FontWeight(600).TextColor(p.OnSurface)

		// Switch 直接绑定设置字段；OnChange 在绑定值写入之后才跑，写库即可。
		sw := ui.Switch(c.Key("backup-switch"), &a.Settings.AutoBackup).
			Label("开启自动备份").Tooltip("每天第一次启动时自动备份一次")
		sw.OnChange(a.SaveSettings)

		ui.Text(c, "每天首次启动时备份").FontSize(design.FontCaption).
			TextColor(p.OnSurfaceVariant)
		ui.Spacer(c)

		// 「保留 N 份」：NumberInput 绑定 float64，本地副本每帧从设置推出。
		keep := float64(a.Settings.BackupKeep)
		ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "保留").FontSize(design.FontCaption).TextColor(p.OnSurfaceVariant)
			num := ui.NumberInput(c.Key("backup-keep"), &keep,
				model.BackupKeepMin, model.BackupKeepMax, 1).Label("保留份数").Width(112)
			num.OnChange(func() {
				a.Settings.BackupKeep = int(keep)
				a.SaveSettings()
			})
			ui.Text(c, "份").FontSize(design.FontCaption).TextColor(p.OnSurfaceVariant)
		})
	})

	line, dir := a.BackupStatus()
	ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
		ui.Text(c, line).FontSize(design.FontCaption).TextColor(p.OnSurfaceVariant).
			SingleLine().Grow(1).Tooltip(dir)
		design.TextButton(c, "backup-now", "立即备份", a.BackupNow)
	})
}

// settingsShortcuts 列出快捷键。
func settingsShortcuts(c *ui.Context) {
	p := design.PaletteOf(c)
	keys := []struct{ Key, What string }{
		{app.CmdName + " + N", "写一条今天的记录"},
		{app.CmdName + " + 1…5", "首页 / 学记 / 随笔 / 待办 / 设置"},
		{app.CmdName + " + F", "在学记里搜索"},
		{app.CmdName + " + Enter", "保存正在写的记录"},
		{"Esc", "关闭弹窗"},
	}
	design.SectionCard(c, "快捷键", "", func() {
		ui.Column(c).Gap(2).Children(func() {
			for _, k := range keys {
				design.KVRow(c, k.Key, func() {
					ui.Text(c, k.What).FontSize(design.FontBody).TextColor(p.OnSurfaceVariant)
				})
			}
		})
	})
}

// settingsAbout 是应用信息与更新。
func settingsAbout(c *ui.Context, a *app.App) {
	p := design.PaletteOf(c)
	design.SectionCard(c, "关于", "", func() {
		ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
			if bmp := app.Icon(); bmp != nil {
				ui.Image(c, bmp).Size(56, 56).Radius(16).Shrink(0)
			}
			ui.Column(c).Gap(4).Children(func() {
				ui.Row(c).Gap(8).AlignItems(ui.Baseline).Children(func() {
					ui.Text(c, app.Name).FontSize(design.FontHero).FontWeight(600).
						TextColor(p.OnSurface)
					ui.Text(c, app.NameEn).FontSize(design.FontSmall).TextColor(p.OnSurfaceVariant)
					design.Chip(c, "v"+app.Version, design.TonePrimary)
				})
				ui.Text(c, app.Tagline).FontSize(design.FontSmall).TextColor(p.OnSurfaceVariant)
			})
		})

		design.Divider(c)

		design.KVRow(c, "作者", func() {
			ui.Text(c, app.Author).FontSize(design.FontBody).TextColor(p.OnSurface)
		})
		design.KVRow(c, "仓库", func() {
			ui.Link(c, app.RepoURL, app.RepoURL).FontSize(design.FontBody).
				TextColor(p.Primary)
		})
		design.KVRow(c, "许可证", func() {
			ui.Text(c, app.License).FontSize(design.FontBody).TextColor(p.OnSurface)
		})
		design.KVRow(c, "界面", func() {
			ui.Text(c, "MyGo "+app.MyGoVersionText()+" · 原生 UI").
				FontSize(design.FontBody).TextColor(p.OnSurface)
		})
		design.KVRow(c, "编译", func() {
			ui.Text(c, "Go "+app.GoVersion()+" · "+app.PlatformName()).
				FontSize(design.FontBody).TextColor(p.OnSurfaceVariant)
		})

		design.Divider(c)
		updateRow(c, a)
	})
}

// updateRow 是「检查更新 / 安装更新」那一块。
func updateRow(c *ui.Context, a *app.App) {
	p := design.PaletteOf(c)

	ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Gap(3).Grow(1).MinWidth(0).Children(func() {
			ui.Text(c, "版本更新").FontSize(design.FontBody).FontWeight(600).TextColor(p.OnSurface)
			switch {
			case !app.UpdateSupported():
				ui.Text(c, "开发版与包管理器安装的版本不支持自动更新").
					FontSize(design.FontCaption).TextColor(p.OnSurfaceVariant)
			case a.Update.Installing:
				ui.Textf(c, "正在下载并安装，%d%%", int(a.Update.Progress*100)).
					FontSize(design.FontCaption).TextColor(p.OnSurfaceVariant)
			case a.Update.Version != "":
				ui.Text(c, "发现新版本 v"+a.Update.Version+"，安装后会自动重启").
					FontSize(design.FontCaption).TextColor(p.OnSurfaceVariant)
			default:
				ui.Text(c, "从 "+app.UpdateHost+" 获取更新").
					FontSize(design.FontCaption).TextColor(p.OnSurfaceVariant)
			}
		})

		switch {
		case !app.UpdateSupported():
			design.OutlinedButton(c, "check-update", "检查更新", nil, nil).Disabled(true)
		case a.Update.Installing:
			// 检查按钮变成「更新中」，进度条只留下面那一条。
			design.BusyButton(c, "更新中")
		case a.Update.Version != "":
			design.FilledButton(c, "install-update", "安装并重启", nil, a.InstallUpdate)
		default:
			label := "检查更新"
			if a.Update.Checking {
				label = "检查中…"
			}
			b := design.OutlinedButton(c, "check-update", label, nil, func() { a.CheckUpdates(true) })
			b.Disabled(a.Update.Checking)
		}
	})

	if a.Update.Installing {
		ui.Progress(c, a.Update.Progress).FillWidth().Height(4)
	}
	if a.Update.Err != "" {
		ui.Text(c, a.Update.Err).FontSize(design.FontCaption).TextColor(p.Error)
	}

	// 有新版本时先看新版本的说明，否则看当前版本更新了什么。
	switch {
	case a.Update.Notes != "" && a.Update.Version != "":
		ui.Text(c, "v"+a.Update.Version+" 更新内容").FontSize(design.FontCaption).
			FontWeight(600).TextColor(p.OnSurfaceVariant)
		ui.Text(c, notesText(a.Update.Notes, 6)).FontSize(design.FontCaption).
			TextColor(p.OnSurfaceVariant).LineHeight(1.65)
	case a.ReleaseNotes != "":
		ui.Text(c, "本次更新（v"+app.Version+"）").FontSize(design.FontCaption).
			FontWeight(600).TextColor(p.OnSurfaceVariant)
		ui.Text(c, notesText(a.ReleaseNotes, 8)).FontSize(design.FontCaption).
			TextColor(p.OnSurfaceVariant).LineHeight(1.65)
	}
}

// notesText 把 CHANGELOG 的 Markdown 列表转成界面上的普通文本。
func notesText(md string, max int) string {
	lines := strings.Split(strings.TrimSpace(md), "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")
		out = append(out, "· "+line)
	}
	if len(out) > max {
		out = append(out[:max], "…")
	}
	return strings.Join(out, "\n")
}

// 让 fmt 在只有格式化字符串的地方也能被用到。
var _ = fmt.Sprintf
