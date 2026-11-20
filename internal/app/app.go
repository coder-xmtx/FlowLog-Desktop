// Package app 是流记的应用状态与动作。
//
// 它管数据：学记、待办、设置，以及所有会改数据的操作（保存、删除、撤销、
// 备份、恢复、检查更新）。它不画界面：页面在 internal/pages，外壳在 internal/view，
// 两者都读这里的字段、调这里的方法。
//
// 状态集中在一个 App 上，界面是它的函数；这样「数据改了什么」只有一处入口，
// 也方便测试直接对状态做断言，不必经过渲染。
package app

import (
	"sort"
	"time"

	"flowlog-desktop/internal/model"
	"flowlog-desktop/internal/store"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// App 是窗口里显示的全部状态。
//
// 字段是导出的：页面直接读，动作方法负责改。写操作一律走方法，
// 不要在页面里直接给切片追加元素——方法里还要落库、去重、排序。
type App struct {
	// Store 是数据库句柄；Win 用于后台任务回到主线程。
	Store *store.Store
	Win   *mygo.Window
	// Router 管页面路径。
	Router *ui.Router

	// Days 有记录的日子，按日期倒序；今天永远在列表里（哪怕还空着）。
	Days []model.Day
	// Todos 全部待办，独立于学记。
	Todos []model.Todo
	// Today 是今天的日期字符串，跨零点时由 RollOver 更新。
	Today string

	// Settings 是应用设置。
	Settings model.Settings
	// LastDark 是上一帧实际用的深浅色，供没有 Context 的切换回调使用。
	LastDark bool

	// ---- 学记页 ----

	// StudyList 是学记列表的滚动与测量状态。
	StudyList ui.ListState
	// Query 是搜索框里的关键词，空字符串表示不搜索。
	Query string
	// SubjectFilter 是按科目筛选时选中的科目，空字符串表示全部科目。
	SubjectFilter string
	// FocusSearch 表示下一帧把焦点送回搜索框。
	FocusSearch bool
	// Filtered 是当前筛选后可见的日子（Days 的下标），每帧重算。
	Filtered []int
	// RevealDate 是侧边栏日历点中的日期：列表建好后滚到那一行。
	RevealDate string

	// ---- 待办页 ----

	// TodoList 是待办列表的滚动与测量状态。
	TodoList ui.ListState
	// TodoDraft 是输入框里的草稿。
	TodoDraft string
	// TodoDue 是新待办的截止日期，零值表示不设。
	TodoDue time.Time
	// CalView 是截止日期弹层里日历的游标。
	CalView time.Time
	// TodoFilter 是 0 全部 / 1 进行中 / 2 已完成。
	TodoFilter int
	// Visible 是当前筛选后可见的待办（Todos 的下标），每帧重算。
	Visible []int
	// FocusInput 表示下一帧把焦点送回待办输入框。
	FocusInput bool

	// CalMonth 是侧边栏月历显示的月份（当月 1 号）。
	CalMonth time.Time

	// ---- 弹层 ----

	// Editor 是「写学记」弹窗的草稿。
	Editor EditorState
	// Confirm 是全局确认框。
	Confirm ConfirmState
	// Onboarding 是首次启动的称呼引导。
	Onboarding OnboardingState
	// DuePicker 表示截止日期弹层是否打开。
	DuePicker bool

	// ---- 首页背景（必应每日一图）----

	// BG 是当日背景图，nil 表示还在加载或取不到。
	BG *ui.Bitmap
	// BGCopyright 是图片版权说明。
	BGCopyright string
	// BGLoading 表示背景图正在下载。
	BGLoading bool

	// ---- 自动更新 ----

	// Update 是自动更新的状态。
	Update UpdateState
	// ReleaseNotes 是当前版本的发布说明（来自 CHANGELOG.md）。
	ReleaseNotes string

	// toasts 是待展示的提示，由界面在帧的开头取走。
	toasts []Toast
}

// EditorState 是「写学记」弹窗的草稿。
type EditorState struct {
	// Open 表示弹窗是否显示。
	Open bool
	// Date 是这条记录所属的日期。
	Date time.Time
	// EntryID 是正在编辑的条目 ID；空字符串表示新增。
	EntryID string
	// Subject、Chapter、Section 是科目与章、节。
	Subject, Chapter, Section string
	// Points 是要点，一行一条。
	Points string
	// Note 是备注。
	Note string
	// Minutes 是学习时长（分钟）。
	Minutes float64
	// Err 是校验失败时的提示。
	Err string
}

// ConfirmState 是一个确认弹窗。
type ConfirmState struct {
	// Open 表示弹窗是否显示。
	Open bool
	// Title 与 Message 是标题与正文。
	Title, Message string
	// Yes 是确认按钮的文字。
	Yes string
	// YesFunc 是确认后的动作。
	YesFunc func()
}

// OnboardingState 是首次启动时的用户名引导。
type OnboardingState struct {
	// Open 表示引导是否显示。
	Open bool
	// Name 是输入框里的称呼。
	Name string
	// Err 是校验提示。
	Err string
	// Focus 表示下一帧把焦点送回输入框。
	Focus bool
}

// UpdateState 是自动更新的进度与结果。
type UpdateState struct {
	// Checking 表示正在检查。
	Checking bool
	// Installing 表示正在下载安装。
	Installing bool
	// Progress 是安装进度 0–1。
	Progress float64
	// Err 是上一次检查或安装的错误。
	Err string
	// Version 是新版本号；空表示已是最新。
	Version string
	// Notes 是新版本的更新说明。
	Notes string
	// available 是待安装的更新，只在主线程读写。
	available *mygo.Update
}

// Toast 是一条待展示的提示。
type Toast struct {
	// Title 是标题，也用作去重 ID。
	Title string
	// Desc 是补充说明。
	Desc string
	// Type 是 "error" 时显示为错误样式。
	Type string
	// Action 是动作按钮文字，OnAct 是点击动作。
	Action string
	OnAct  func()
}

// New 载入数据库，建立应用状态。
//
// 参数 s 是已经打开的数据库；返回的状态里已经包含今天这一天。
func New(s *store.Store) (*App, error) {
	days, err := s.Days()
	if err != nil {
		return nil, err
	}
	todos, err := s.Todos()
	if err != nil {
		return nil, err
	}
	settings, err := s.Settings()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	a := &App{
		Store:    s,
		Router:   ui.NewRouter("/"),
		Today:    model.Today(),
		Days:     days,
		Todos:    todos,
		Settings: settings,
		CalMonth: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local),
		CalView:  now,
	}
	a.EnsureToday()
	return a, nil
}

// Start 做那些要等窗口就绪之后才能开始的事：引导、背景图、备份、更新检查。
func (a *App) Start() {
	if !a.Settings.Onboarded {
		a.Onboarding.Open = true
		a.Onboarding.Name = a.Settings.UserName
	}
	a.LoadBackground()
	a.AutoBackup()
	a.CheckForUpdatesLater()
}

// SetWindow 记住主窗口，供后台任务请求重绘、文件对话框使用。
func (a *App) SetWindow(w *mygo.Window) { a.Win = w }

// SetReleaseNotes 传入 CHANGELOG.md 的内容，设置页会展示当前版本的说明。
func (a *App) SetReleaseNotes(changelog string) {
	a.ReleaseNotes = ExtractReleaseNotes(changelog, Version)
}

// ---------------------------------------------------------------- 日期与学记

// DayIndex 返回某一天在 Days 里的下标，没有为 -1。
func (a *App) DayIndex(date string) int {
	for i := range a.Days {
		if a.Days[i].Date == date {
			return i
		}
	}
	return -1
}

// Day 返回某一天的记录；不存在时返回一个空 Day（不写库）。
func (a *App) Day(date string) model.Day {
	if i := a.DayIndex(date); i >= 0 {
		return a.Days[i]
	}
	return model.Day{Date: date, Entries: []model.Entry{}}
}

// EnsureToday 保证今天在列表里，这样「今天」永远有一张可以落笔的卡片。
func (a *App) EnsureToday() {
	if a.DayIndex(a.Today) >= 0 {
		return
	}
	pos := sort.Search(len(a.Days), func(i int) bool { return a.Days[i].Date <= a.Today })
	a.Days = append(a.Days, model.Day{})
	copy(a.Days[pos+1:], a.Days[pos:])
	a.Days[pos] = model.Day{Date: a.Today, Entries: []model.Entry{}}
}

// RollOver 处理跨零点：应用一直开着的时候，「今天」要跟着日期走。
func (a *App) RollOver() {
	today := model.Today()
	if today == a.Today {
		return
	}
	a.Today = today
	a.EnsureToday()
}

// Upsert 把一天写进内存列表，并保持按日期倒序。
func (a *App) Upsert(day model.Day) {
	if i := a.DayIndex(day.Date); i >= 0 {
		a.Days[i] = day
		return
	}
	pos := 0
	for pos < len(a.Days) && a.Days[pos].Date > day.Date {
		pos++
	}
	a.Days = append(a.Days, model.Day{})
	copy(a.Days[pos+1:], a.Days[pos:])
	a.Days[pos] = day
}

// Persist 把某一天的记录写进数据库，并清掉已经空掉的日子。
//
// 参数 date 是这一天的 "2006-01-02"；写入失败会弹一条错误提示。
func (a *App) Persist(date string) {
	i := a.DayIndex(date)
	if i < 0 {
		return
	}
	if err := a.Store.SaveDay(a.Days[i]); err != nil {
		a.ToastError("保存失败", err.Error())
		return
	}
	a.PruneEmpty()
}

// PruneEmpty 丢掉「已经不是今天、且没有学习记录」的日子。
func (a *App) PruneEmpty() {
	kept := a.Days[:0]
	for _, d := range a.Days {
		if d.HasEntries() || d.Date == a.Today {
			kept = append(kept, d)
		}
	}
	a.Days = kept
}

// SaveSettings 把设置写进数据库。
func (a *App) SaveSettings() {
	if err := a.Store.SaveSettings(a.Settings); err != nil {
		a.ToastError("保存设置失败", err.Error())
	}
}

// ---------------------------------------------------------------- 统计

// TodayEntryCount 返回今天写了几条学习记录。
func (a *App) TodayEntryCount() int {
	i := a.DayIndex(a.Today)
	if i < 0 {
		return 0
	}
	return len(a.Days[i].Entries)
}

// OpenTodoCount 返回全部未完成的待办数。
func (a *App) OpenTodoCount() int {
	n := 0
	for _, t := range a.Todos {
		if !t.Done {
			n++
		}
	}
	return n
}

// DaysWritten 返回写过学记的天数。
func (a *App) DaysWritten() int {
	n := 0
	for i := range a.Days {
		if a.Days[i].HasEntries() {
			n++
		}
	}
	return n
}

// Summary 汇总全库的规模，设置页会用。
func (a *App) Summary() (days, entries, points, todos int) {
	for _, d := range a.Days {
		if !d.HasEntries() {
			continue
		}
		days++
		entries += len(d.Entries)
		points += d.PointCount()
	}
	return days, entries, points, len(a.Todos)
}

// HasEntries 报告某天有没有记录，侧边栏日历用它打点。
func (a *App) HasEntries(date string) bool {
	i := a.DayIndex(date)
	return i >= 0 && a.Days[i].HasEntries()
}

// Streak 计算连续记录的天数。
//
// 今天还没写不算断：从昨天往前数，昨天写过就仍然连续。
func (a *App) Streak() int {
	day := model.MustDate(a.Today)
	if day.IsZero() {
		return 0
	}
	if !a.HasEntries(a.Today) {
		day = day.AddDate(0, 0, -1)
	}
	n := 0
	for ; n < 3650; n++ {
		if !a.HasEntries(day.Format(model.DateLayout)) {
			break
		}
		day = day.AddDate(0, 0, -1)
	}
	return n
}

// ---------------------------------------------------------------- 提示

// Toast 是一条普通提示。
func (a *App) Toast(title, desc string) {
	a.toasts = append(a.toasts, Toast{Title: title, Desc: desc})
}

// ToastError 是一条错误提示。
func (a *App) ToastError(title, desc string) {
	a.toasts = append(a.toasts, Toast{Title: title, Desc: desc, Type: "error"})
}

// ToastUndo 是一条可撤销的提示。
//
// 参数 action 是动作按钮的文字，undo 是点它以后执行的动作。
func (a *App) ToastUndo(title, action string, undo func()) {
	a.toasts = append(a.toasts, Toast{Title: title, Action: action, OnAct: undo})
}

// TakeToasts 取走攒下的提示，交给界面在帧的开头展示。
func (a *App) TakeToasts() []Toast {
	out := a.toasts
	a.toasts = nil
	return out
}

// Redraw 在后台拿到新数据后请求一帧。
func (a *App) Redraw() {
	if a.Win != nil {
		a.Win.Invalidate()
	}
}

// Post 把 fn 放到主线程执行并重绘；没有窗口时直接执行。
//
// 后台 goroutine 改界面状态（包括弹提示）一律经过它。
func (a *App) Post(fn func()) {
	if a.Win != nil {
		a.Win.Update(fn)
		return
	}
	fn()
}
