package view

import (
	"strings"

	"flowlog-desktop/internal/app"
	design "flowlog-desktop/internal/design"

	"github.com/egoist/mygo/ui"
)

// ConfirmDialog 是全局的确认框：删除记录、恢复备份都用它。
func ConfirmDialog(c *ui.Context, a *app.App) {
	if !a.Confirm.Open {
		return
	}
	st := &a.Confirm
	design.Confirm(c, &st.Open, st.Title, st.Message, st.Yes, st.YesFunc)
}

// OnboardingDialog 是首次启动时询问称呼的弹窗。
//
// 用户关掉它（Esc / 点外面）也算完成引导：设置里随时能改。
func OnboardingDialog(c *ui.Context, a *app.App) {
	if !a.Onboarding.Open {
		// 引导已被关掉：标记完成，下次不再打扰。
		if !a.Settings.Onboarded {
			a.Settings.Onboarded = true
			a.SaveSettings()
		}
		return
	}

	p := design.PaletteOf(c)
	design.Dialog(c, &a.Onboarding.Open, func() {
		ui.Column(c).Width(420).Gap(16).Children(func() {
			ui.Column(c).Gap(6).Children(func() {
				design.Text(c, "先认识一下").FontSize(design.FontHero).FontWeight(600).
					TextColor(p.OnSurface)
				design.Text(c, "留下一个称呼，首页会用它问候你，之后可以在设置里修改").
					FontSize(design.FontBody).TextColor(p.OnSurfaceVariant).LineHeight(1.6)
			})

			ui.Field(c, "称呼", func() {
				in := ui.TextInput(c.Key("onboarding-name"), &a.Onboarding.Name).
					Label("称呼").AutoFocus().MaxLength(24).
					Placeholder("例如：小明")
				if a.Onboarding.Focus {
					// 校验失败后把焦点送回输入框，用户可以直接接着打字。
					in.Focus()
					a.Onboarding.Focus = false
				}
			})

			if a.Onboarding.Err != "" {
				design.Text(c, a.Onboarding.Err).FontSize(design.FontSmall).TextColor(p.Error)
			}

			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				design.TextButton(c, "onboarding-skip", "跳过", a.SkipOnboarding)
				ui.Spacer(c)
				design.FilledButton(c, "onboarding-start", "开始使用", nil, a.FinishOnboarding)
			})
		})
	})

	if a.Onboarding.Open && c.Shortcut(0, ui.KeyEnter) {
		a.FinishOnboarding()
	}
}

// joinNames 把文件名列表拼成一行提示。
func joinNames(names []string) string {
	if len(names) > 3 {
		return strings.Join(names[:3], "、") + " 等"
	}
	return strings.Join(names, "、")
}
