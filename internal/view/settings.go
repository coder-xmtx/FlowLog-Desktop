package view

import (
	"strings"

	"flowlog-desktop/internal/model"

	"github.com/egoist/mygo/ui"
)

// onboardingDialog 是首次启动时询问称呼的弹窗。
//
// 用户关掉它（Esc / 点外面）也算完成引导：设置里随时能改。
func (app *App) onboardingDialog(c *ui.Context) {
	if !app.onboarding.Open {
		// 引导已被关掉：标记完成，下次不再打扰。
		if !app.settings.Onboarded {
			app.settings.Onboarded = true
			app.saveSettings()
		}
		return
	}

	t := c.Theme()
	ui.Modal(c, &app.onboarding.Open, func() {
		ui.Column(c).Width(420).Gap(16).Children(func() {
			ui.Column(c).Gap(6).Children(func() {
				ui.Text(c, "先认识一下").Font(fontSerif).FontSize(20).Bold()
				ui.Text(c, "留下一个称呼，首页会用它问候你。之后可以在设置里修改。").
					FontSize(12.5).TextColor(t.TextMuted).LineHeight(1.6)
			})

			ui.Field(c, "称呼", func() {
				in := ui.TextInput(c.Key("onboarding-name"), &app.onboarding.Name).
					Label("称呼").AutoFocus().MaxLength(24).
					Placeholder("例如：小明")
				if app.onboarding.Focus {
					// 校验失败后把焦点送回输入框，用户可以直接接着打字。
					in.Focus()
					app.onboarding.Focus = false
				}
			})

			if app.onboarding.Err != "" {
				ui.Text(c, app.onboarding.Err).FontSize(12).TextColor(t.Danger)
			}

			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Button(c, "跳过").OnClick(app.skipOnboarding)
				ui.Spacer(c)
				ui.PrimaryButton(c, "开始使用").OnClick(app.finishOnboarding)
			})
		})
	})

	if app.onboarding.Open && c.Shortcut(0, ui.KeyEnter) {
		app.finishOnboarding()
	}
}

// finishOnboarding 保存称呼并结束引导。
func (app *App) finishOnboarding() {
	name := strings.TrimSpace(app.onboarding.Name)
	if name == "" {
		app.onboarding.Err = "写一个称呼，或者点「跳过」。"
		app.onboarding.Focus = true
		return
	}
	app.settings.UserName = name
	app.settings.Onboarded = true
	app.saveSettings()
	app.onboarding.Open = false
	app.toast("欢迎使用流记", "你好，"+name)
}

// skipOnboarding 跳过引导，不保存称呼。
func (app *App) skipOnboarding() {
	app.settings.Onboarded = true
	app.onboarding.Open = false
	app.saveSettings()
}

// greetingName 返回称呼前的问候语。
func (app *App) greetingName() string {
	if name := strings.TrimSpace(app.settings.UserName); name != "" {
		return name
	}
	return ""
}

// settingsPage 是「设置」页：先设置，后关于。
func (app *App) settingsPage(c *ui.Context) {
	days, entries, points, todos := app.summary()

	ui.Column(c).Grow(1).MinWidth(0).Children(func() {
		pageHeader(c, "设置", "外观、账户与数据；关于本应用的信息在最后。", nil)

		ui.Scroll(c).Grow(1).Padding(4, 32, 28).Gap(16).Children(func() {
			ui.Column(c).MaxWidth(680).Gap(16).Children(func() {
				app.settingsAccount(c)
				app.settingsAppearance(c)
				app.settingsData(c, days, entries, points, todos)
				app.settingsShortcuts(c)
				app.settingsAbout(c)
			})
		})
	})
}

// settingsAccount 是称呼。
func (app *App) settingsAccount(c *ui.Context) {
	t := c.Theme()
	sectionCard(c, "账户", "", func() {
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			in := ui.TextInput(c.Key("settings-name"), &app.settings.UserName).
				Label("称呼").MaxLength(24).Placeholder("你的称呼").Grow(1)
			if in.Submitted() {
				app.saveSettings()
				app.toast("已保存", "称呼已更新")
			}
			ui.Button(c, "保存").OnClick(func() {
				app.saveSettings()
				app.toast("已保存", "称呼已更新")
			})
		})
		ui.Text(c, "首页的问候会用到这个称呼，随时可以修改。").
			FontSize(11.5).TextColor(t.TextMuted)
	})
}

// settingsAppearance 是深色 / 浅色 / 跟随系统。
func (app *App) settingsAppearance(c *ui.Context) {
	t := c.Theme()
	sectionCard(c, "外观", "", func() {
		// Segmented 绑定的是下标。控件在构造返回时就写入了用户的点击，
		// 所以 OnChange 里读到的 choice 已经是新值。
		choice := 0
		switch app.settings.Theme {
		case model.ThemeLight:
			choice = 1
		case model.ThemeSystem:
			choice = 2
		}
		ui.Segmented(c.Key("settings-theme"), &choice, "深色", "浅色", "跟随系统").
			Label("外观").FillEqually().
			OnChange(func() {
				theme := model.ThemeDark
				switch choice {
				case 1:
					theme = model.ThemeLight
				case 2:
					theme = model.ThemeSystem
				}
				if theme != app.settings.Theme {
					app.settings.Theme = theme
					app.saveSettings()
				}
			})
		ui.Text(c, "默认深色；选择「跟随系统」后，应用会随系统的深浅色切换。").
			FontSize(11.5).TextColor(t.TextMuted)
	})
}

// settingsData 是数据库信息与备份、迁移。
func (app *App) settingsData(c *ui.Context, days, entries, points, todos int) {
	t := c.Theme()
	path := app.store.Path()

	sectionCard(c, "数据", "本地 SQLite 数据库", func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Textf(c, "%d 天 · %d 条记录 · %d 条要点 · %d 条待办",
				days, entries, points, todos).FontSize(12.5)
			ui.Spacer(c)
			ui.Text(c, formatSize(app.store.Size())).FontSize(11.5).TextColor(t.TextMuted)
		})

		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Grow(1).Padding(7, 10).Radius(7).Background(t.Surface).Children(func() {
				ui.Text(c, path).FontSize(11.5).Font("monospace").Selectable().SingleLine().
					Tooltip(path)
			})
			iconTextButton(c, "copy-db-path", "复制路径", iconCopy, func() {
				c.WriteClipboard(path)
				app.toast("已复制路径", path)
			})
		})

		ui.Row(c).Gap(8).Wrap().Children(func() {
			iconTextButton(c, "export-backup", "导出备份", iconDownload, app.exportBackup)
			iconTextButton(c, "restore-backup", "从备份恢复", iconUpload, app.chooseBackup)
		})

		ui.Text(c, "备份是一个 JSON 文件，包含全部日记、待办与设置，可以复制到别的机器上恢复。").
			FontSize(11.5).TextColor(t.TextMuted).LineHeight(1.6)
	})
}

// settingsShortcuts 列出快捷键。
func (app *App) settingsShortcuts(c *ui.Context) {
	keys := []struct{ Key, What string }{
		{cmdName + " + N", "写一条今天的记录"},
		{cmdName + " + 1…4", "首页 / 日记 / 待办 / 设置"},
		{cmdName + " + F", "在日记里搜索"},
		{cmdName + " + Enter", "保存正在写的日记"},
		{"Esc", "关闭弹窗"},
	}
	sectionCard(c, "快捷键", "", func() {
		ui.Column(c).Gap(2).Children(func() {
			for _, k := range keys {
				kvRow(c, k.Key, func() {
					ui.Text(c, k.What).FontSize(12.5).TextColor(c.Theme().TextMuted)
				})
			}
		})
	})
}

// settingsAbout 是应用信息与更新。
func (app *App) settingsAbout(c *ui.Context) {
	t := c.Theme()
	sectionCard(c, "关于", "", func() {
		ui.Row(c).Gap(16).AlignItems(ui.Center).Children(func() {
			if bmp := iconBitmap(); bmp != nil {
				ui.Image(c, bmp).Size(56, 56).Radius(14).Shrink(0)
			}
			ui.Column(c).Gap(3).Children(func() {
				ui.Row(c).Gap(8).AlignItems(ui.Baseline).Children(func() {
					ui.Text(c, appName).Font(fontSerif).FontSize(20).Bold()
					ui.Text(c, appNameEn).FontSize(12).TextColor(t.TextMuted)
					pill(c, "v"+appVersion, t.Accent.Alpha(0.13), t.Accent)
				})
				ui.Text(c, appTagline).FontSize(12).TextColor(t.TextMuted)
			})
		})

		ui.Box(c).Height(1).Background(t.Border)

		kvRow(c, "作者", func() { ui.Text(c, appAuthor).FontSize(12.5) })
		kvRow(c, "仓库", func() {
			ui.Link(c, appRepoURL, appRepoURL).FontSize(12.5)
		})
		kvRow(c, "许可证", func() { ui.Text(c, appLicense).FontSize(12.5) })
		kvRow(c, "界面", func() {
			ui.Text(c, "MyGo "+mygoVersionText()+" · 原生 UI（非 WebView）").FontSize(12.5)
		})
		kvRow(c, "编译", func() {
			ui.Text(c, "Go "+goVersion()+" · "+platformName()).FontSize(12.5).
				TextColor(t.TextMuted)
		})

		ui.Box(c).Height(1).Background(t.Border)
		app.updateRow(c)
	})
}

// updateRow 是「检查更新 / 安装更新」那一行。
func (app *App) updateRow(c *ui.Context) {
	t := c.Theme()

	ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
		ui.Column(c).Gap(2).Grow(1).MinWidth(0).Children(func() {
			ui.Text(c, "版本更新").FontSize(12.5).FontWeight(600)
			switch {
			case !updateSupported():
				ui.Text(c, "开发版与包管理器安装的版本不支持自动更新。").
					FontSize(11.5).TextColor(t.TextMuted)
			case app.update.Installing:
				ui.Textf(c, "正在下载并安装，%d%%", int(app.update.Progress*100)).
					FontSize(11.5).TextColor(t.TextMuted)
			case app.update.Version != "":
				ui.Text(c, "发现新版本 v"+app.update.Version+"，安装后会自动重启。").
					FontSize(11.5).TextColor(t.TextMuted)
			default:
				ui.Text(c, "从 GitHub Releases 获取更新。").
					FontSize(11.5).TextColor(t.TextMuted)
			}
		})

		switch {
		case !updateSupported():
			ui.Button(c, "检查更新").Disabled(true)
		case app.update.Installing:
			ui.Progress(c, app.update.Progress).Width(120)
		case app.update.Version != "":
			ui.PrimaryButton(c, "安装并重启").OnClick(app.installUpdate)
		default:
			label := "检查更新"
			if app.update.Checking {
				label = "检查中…"
			}
			ui.Button(c, label).Disabled(app.update.Checking).
				OnClick(func() { app.checkUpdates(true) })
		}
	})

	if app.update.Installing {
		ui.Progress(c, app.update.Progress).FillWidth().Height(4)
	}
	if app.update.Err != "" {
		ui.Text(c, app.update.Err).FontSize(11.5).TextColor(t.Danger)
	}

	// 有新版本时先看新版本的说明，否则看当前版本更新了什么。
	switch {
	case app.update.Notes != "" && app.update.Version != "":
		ui.Text(c, "v"+app.update.Version+" 更新内容").FontSize(11.5).
			FontWeight(600).TextColor(t.TextMuted)
		ui.Text(c, notesText(app.update.Notes, 6)).FontSize(11.5).
			TextColor(t.TextMuted).LineHeight(1.65)
	case app.releaseNotes != "":
		ui.Text(c, "本次更新（v"+appVersion+"）").FontSize(11.5).
			FontWeight(600).TextColor(t.TextMuted)
		ui.Text(c, notesText(app.releaseNotes, 8)).FontSize(11.5).
			TextColor(t.TextMuted).LineHeight(1.65)
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

// iconTextButton 是「图标 + 文字」的次要按钮。
func iconTextButton(c *ui.Context, key any, label string, icon *ui.SVG, on func()) ui.Element {
	b := ui.Button(c.Key(key), "").Label(label).FontSize(12)
	b.Children(func() {
		ui.Icon(c, icon).FontSize(13)
		ui.Text(c, label).SingleLine()
	})
	return b.OnClick(on)
}
