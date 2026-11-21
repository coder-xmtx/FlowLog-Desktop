package app

import (
	"flowlog-desktop/internal/lang"
	"github.com/egoist/mygo"
	"strings"

	"flowlog-desktop/internal/model"
)

// 设置相关的动作：称呼引导、外观切换、问候语。
//
// 具体怎么画在 internal/pages/settings.go 与 internal/view 的引导弹窗里。

// FinishOnboarding 保存称呼并结束引导。
func (a *App) FinishOnboarding() {
	name := strings.TrimSpace(a.Onboarding.Name)
	if name == "" {
		a.Onboarding.Err = "写一个称呼，或者点「跳过」。"
		a.Onboarding.Focus = true
		return
	}
	a.Settings.UserName = name
	a.Settings.Onboarded = true
	a.SaveSettings()
	a.Onboarding.Open = false
	a.Toast(lang.T("欢迎使用流记"), lang.Format("你好，%s", name))
}

// SkipOnboarding 跳过引导：不写称呼，但记住已经问过了。
func (a *App) SkipOnboarding() {
	a.Settings.Onboarded = true
	a.SaveSettings()
	a.Onboarding.Open = false
}

// GreetingName 返回可以用来问候的称呼，没设置时返回空字符串。
func (a *App) GreetingName() string {
	return strings.TrimSpace(a.Settings.UserName)
}

// ToggleTheme 在浅色与深色之间切换，并写库。
//
// 「跟随系统」时按当前实际显示的那一套取反。
func (a *App) ToggleTheme() {
	dark := a.EffectiveDarkNow()
	if dark {
		a.Settings.Theme = model.ThemeLight
	} else {
		a.Settings.Theme = model.ThemeDark
	}
	a.SaveSettings()
}

// EffectiveDarkNow 是给没有 Context 的调用点用的近似值：
// 「跟随系统」时读上一帧实际用的深浅色（LastDark）。
func (a *App) EffectiveDarkNow() bool {
	switch a.Settings.Theme {
	case model.ThemeLight:
		return false
	case model.ThemeDark:
		return true
	}
	return a.LastDark
}

// SetTheme 设置外观并落库（"dark" / "light" / "system"）。
func (a *App) SetTheme(theme string) {
	theme = model.NormalizeTheme(theme)
	if theme == a.Settings.Theme {
		return
	}
	a.Settings.Theme = theme
	a.SaveSettings()
}

// SetLanguage 切换界面语言：写进设置、立即换掉当前语言，
// 并让 MyGo 的内置控件（日历、日期框等）跟着换。
func (a *App) SetLanguage(tag string) {
	tag = lang.Normalize(tag)
	if tag == a.Settings.Language {
		return
	}
	a.Settings.Language = tag
	a.SaveSettings()
	lang.Set(tag)
	mygo.App.SetLocale(tag)
	a.Router.Replace(a.Router.Path())
}

// ApplyLanguage 在启动时应用保存的语言。
func (a *App) ApplyLanguage() {
	tag := lang.Normalize(a.Settings.Language)
	a.Settings.Language = tag
	lang.Set(tag)
	mygo.App.SetLocale(tag)
}
