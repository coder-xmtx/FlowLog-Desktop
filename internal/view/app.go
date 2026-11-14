// Package view 是流记的界面层：状态、路由、四个页面与各种弹层。
//
// 这一层遵守 MyGo 原生 UI 的三条硬规则：
//  1. 视图是「状态的函数」，每帧重建一次；Element 只在当帧有效，
//     绝不能存进 App 或后台 goroutine。
//  2. 会改动「正在被遍历的集合」的动作交给 OnClick / OnChange 延迟执行，
//     它们在本帧构建完之后运行，下一帧重建时读到新状态。
//  3. 控件的状态跟着它的位置或 Key 走，所以列表里的控件要用稳定 Key。
package view

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"flowlog-desktop/internal/model"
	"flowlog-desktop/internal/store"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// App 是窗口里显示的全部状态。
//
// 日记与待办是两套独立的数据：days 只有学习记录，todos 不属于任何一天。
type App struct {
	store  *store.Store
	router *ui.Router
	win    *mygo.Window

	// days 有日记的日子，按日期倒序；今天永远在列表里（哪怕还空着）。
	days []model.Day
	// todos 全部待办，独立于日记。
	todos []model.Todo
	today string

	settings model.Settings
	// lastDark 是上一帧实际用的深浅色，供没有 Context 的切换回调使用。
	lastDark bool

	// 日记页
	diaryList   ui.ListState
	query       string
	focusSearch bool
	filtered    []int
	// revealDate 是侧边栏日历点中的日期：日记页构建完成后滚到那一行。
	revealDate string

	// 待办页
	todoList   ui.ListState
	todoDraft  string
	todoDue    time.Time // 新待办的截止日期，零值表示不设
	calView    time.Time // 截止日期弹层里日历的游标
	focusInput bool      // 下一帧把焦点送回待办输入框
	todoFilter int
	visible    []int // 每帧算出的待办下标

	// 侧边栏日历显示的月份（当月 1 号）。
	calMonth time.Time

	// 弹层
	ed         editorState
	confirm    confirmState
	onboarding onboardState
	duePicker  bool

	// 首页背景（必应每日一图）
	bg          *ui.Bitmap
	bgCopyright string
	bgLoading   bool

	// 自动更新
	update updateState
	// releaseNotes 是当前版本的发布说明（来自 CHANGELOG.md）。
	releaseNotes string

	// 待展示的提示，在下一帧开头交给 c.AddToast。
	toasts []pendingToast
}

// editorState 是「写日记」弹窗的草稿。
type editorState struct {
	Open    bool
	Date    time.Time
	EntryID string // 正在编辑的条目 ID；空字符串表示新增
	Subject string
	Chapter string
	Section string
	Points  string // 一行一条
	Note    string
	Minutes float64
	Err     string
}

// confirmState 是一个确认弹窗。
type confirmState struct {
	Open    bool
	Title   string
	Message string
	Yes     string
	YesFunc func()
}

// onboardState 是首次启动时的用户名引导。
type onboardState struct {
	Open bool
	Name string
	Err  string
	// Focus 表示下一帧要把焦点送回输入框（校验失败后用）。
	Focus bool
}

// pendingToast 是一条等到下一帧再交给 MyGo 的提示。
type pendingToast struct {
	Title  string
	Desc   string
	Type   string
	Action string
	OnAct  func()
}

// NewApp 载入数据库，建立视图状态。
func NewApp(s *store.Store) (*App, error) {
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
	app := &App{
		store:    s,
		router:   ui.NewRouter("/"),
		today:    model.Today(),
		days:     days,
		todos:    todos,
		settings: settings,
		calMonth: time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local),
		calView:  now,
	}
	app.ensureToday()
	return app, nil
}

// Start 做那些要等窗口就绪之后才能开始的事：背景图、更新检查。
func (app *App) Start() {
	if !app.settings.Onboarded {
		app.onboarding.Open = true
		app.onboarding.Name = app.settings.UserName
	}
	app.loadBackground()
	app.checkForUpdatesLater()
}

// SetWindow 记住主窗口，供后台任务请求重绘、菜单项使用。
func (app *App) SetWindow(w *mygo.Window) { app.win = w }

// SetReleaseNotes 传入 CHANGELOG.md 的内容，设置页会展示当前版本的说明。
func (app *App) SetReleaseNotes(changelog string) {
	app.releaseNotes = extractReleaseNotes(changelog, appVersion)
}

// View 是窗口的内容函数：每帧被调用一次（或多次），把状态画成界面。
func (app *App) View(c *ui.Context) {
	app.rollOver()
	app.applyTheme(c)
	app.flushToasts(c)
	app.shortcuts(c)

	root := ui.Row(c).Fill().AlignItems(ui.Stretch)
	root.Children(func() {
		app.sidebar(c)
		app.router.View(c, func(r *ui.Route) {
			switch {
			case r.Match("/"):
				r.Title("流记 · 首页")
				app.home(c)
			case r.Match("/diary"):
				r.Title("流记 · 日记")
				app.diary(c)
			case r.Match("/todo"):
				r.Title("流记 · 待办")
				app.todo(c)
			case r.Match("/settings"):
				r.Title("流记 · 设置")
				app.settingsPage(c)
			default:
				ui.Text(c, "没有这个页面")
			}
		})
	})

	// 弹层写在路由外面：切换页面时它们不会跟着消失。
	app.diaryEditor(c)
	app.confirmDialog(c)
	app.onboardingDialog(c)
}

// ---------------------------------------------------------------- 数据读写

// dayIndex 返回某一天在 app.days 里的下标，没有为 -1。
func (app *App) dayIndex(date string) int {
	for i := range app.days {
		if app.days[i].Date == date {
			return i
		}
	}
	return -1
}

// day 返回某一天的日记；不存在时返回一个空 Day（不写库）。
func (app *App) day(date string) model.Day {
	if i := app.dayIndex(date); i >= 0 {
		return app.days[i]
	}
	return model.Day{Date: date, Entries: []model.Entry{}}
}

// ensureToday 保证今天在列表里，这样「今天」永远有一张可以落笔的卡片。
func (app *App) ensureToday() {
	if app.dayIndex(app.today) >= 0 {
		return
	}
	pos := sort.Search(len(app.days), func(i int) bool { return app.days[i].Date <= app.today })
	app.days = append(app.days, model.Day{})
	copy(app.days[pos+1:], app.days[pos:])
	app.days[pos] = model.Day{Date: app.today, Entries: []model.Entry{}}
}

// rollOver 处理跨零点：应用一直开着的时候，「今天」要跟着日期走。
func (app *App) rollOver() {
	today := model.Today()
	if today == app.today {
		return
	}
	app.today = today
	app.ensureToday()
}

// persist 把某一天的日记写进数据库。
func (app *App) persist(date string) {
	i := app.dayIndex(date)
	if i < 0 {
		return
	}
	if err := app.store.SaveDay(app.days[i]); err != nil {
		app.toastError("保存失败", err.Error())
		return
	}
	app.pruneEmpty()
}

// saveTodo 写入或更新一条待办，并同步内存里的副本。
func (app *App) saveTodo(todo model.Todo) bool {
	if err := app.store.SaveTodo(todo); err != nil {
		app.toastError("保存待办失败", err.Error())
		return false
	}
	if i := app.todoIndex(todo.ID); i >= 0 {
		app.todos[i] = todo
	} else {
		app.todos = append(app.todos, todo)
	}
	app.sortTodos()
	return true
}

// deleteTodoRow 从数据库与内存里删掉一条待办。
func (app *App) deleteTodoRow(id string) bool {
	if err := app.store.DeleteTodo(id); err != nil {
		app.toastError("删除待办失败", err.Error())
		return false
	}
	for i := range app.todos {
		if app.todos[i].ID == id {
			app.todos = append(app.todos[:i], app.todos[i+1:]...)
			return true
		}
	}
	return true
}

// todoIndex 返回某条待办在 app.todos 里的下标。
func (app *App) todoIndex(id string) int {
	for i := range app.todos {
		if app.todos[i].ID == id {
			return i
		}
	}
	return -1
}

// sortTodos 保持与数据库一致的顺序：未完成在前，按截止日期与创建时间。
func (app *App) sortTodos() {
	sort.SliceStable(app.todos, func(i, j int) bool {
		a, b := app.todos[i], app.todos[j]
		if a.Done != b.Done {
			return !a.Done
		}
		if a.Due != b.Due {
			if a.Due == "" {
				return false
			}
			if b.Due == "" {
				return true
			}
			return a.Due < b.Due
		}
		return a.Created > b.Created
	})
}

// pruneEmpty 丢掉「已经不是今天、且没有学习记录」的日子。
func (app *App) pruneEmpty() {
	kept := app.days[:0]
	for _, d := range app.days {
		if d.HasEntries() || d.Date == app.today {
			kept = append(kept, d)
		}
	}
	app.days = kept
}

// saveSettings 把设置写进数据库。
func (app *App) saveSettings() {
	if err := app.store.SaveSettings(app.settings); err != nil {
		app.toastError("保存设置失败", err.Error())
	}
}

// upsert 把一天写进内存列表，并保持按日期倒序。
func (app *App) upsert(day model.Day) {
	if i := app.dayIndex(day.Date); i >= 0 {
		app.days[i] = day
		return
	}
	pos := 0
	for pos < len(app.days) && app.days[pos].Date > day.Date {
		pos++
	}
	app.days = append(app.days, model.Day{})
	copy(app.days[pos+1:], app.days[pos:])
	app.days[pos] = day
}

// ---------------------------------------------------------------- 提示

func (app *App) toast(title, desc string) {
	app.toasts = append(app.toasts, pendingToast{Title: title, Desc: desc})
}

func (app *App) toastError(title, desc string) {
	app.toasts = append(app.toasts, pendingToast{Title: title, Desc: desc, Type: "error"})
}

// toastUndo 是一条可撤销的提示。
func (app *App) toastUndo(title, action string, undo func()) {
	app.toasts = append(app.toasts, pendingToast{Title: title, Action: action, OnAct: undo})
}

// flushToasts 在帧的开头把攒下的提示交给窗口。
func (app *App) flushToasts(c *ui.Context) {
	for _, t := range app.toasts {
		c.AddToast(ui.Toast{
			ID:          t.Title, // 同一个标题的提示就地更新，而不是叠一摞
			Title:       t.Title,
			Description: t.Desc,
			Type:        t.Type,
			Action:      t.Action,
			OnAction:    t.OnAct,
			Timeout:     4 * time.Second,
		})
	}
	app.toasts = app.toasts[:0]
}

// ---------------------------------------------------------------- 快捷键

// shortcuts 处理窗口级快捷键。弹窗开着时 c.Shortcut 自动失效。
func (app *App) shortcuts(c *ui.Context) {
	switch {
	case c.Shortcut(ui.Cmd, ui.KeyN):
		app.openEditor(app.today, "")
	case c.Shortcut(ui.Cmd, ui.Key1):
		app.router.Push("/")
	case c.Shortcut(ui.Cmd, ui.Key2):
		app.router.Push("/diary")
	case c.Shortcut(ui.Cmd, ui.Key3):
		app.router.Push("/todo")
	case c.Shortcut(ui.Cmd, ui.Key4):
		app.router.Push("/settings")
	case c.Shortcut(ui.Cmd, ui.KeyF):
		app.router.Push("/diary")
		app.focusSearch = true
	}
}

// ---------------------------------------------------------------- 侧边栏

// sidebar 是左侧导航：品牌、四个页面、月历、外观与版本。
func (app *App) sidebar(c *ui.Context) {
	t := c.Theme()
	side := ui.Column(c).Width(232).Shrink(0).FillHeight().
		Background(t.Surface).BorderWidth(0, 1, 0, 0).BorderColor(t.Border)

	side.Children(func() {
		app.brand(c)

		page := app.router.Path()
		if ui.Sidebar(c, &page, func() {
			ui.SidebarItem(c, "/", iconHome, "首页")
			// 日记徽标只数今天的记录；累计量放在日历和日记页里看。
			item := ui.SidebarItem(c, "/diary", iconBook, "日记")
			if n := app.todayEntryCount(); n > 0 {
				item.Children(func() { ui.Badge(c, strconv.Itoa(n)) })
			}
			todo := ui.SidebarItem(c, "/todo", iconCheckCircle, "待办")
			if open := app.openTodoCount(); open > 0 {
				todo.Children(func() { ui.Badge(c, strconv.Itoa(open)) })
			}
			ui.SidebarItem(c, "/settings", iconSettings, "设置")
		}).Grow(1).Label("主导航").Changed() {
			app.router.Push(page)
		}

		app.monthCalendar(c)
		app.sidebarFooter(c)
	})
}

// brand 是侧边栏顶部的标志：直接用应用图标。
func (app *App) brand(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).Gap(10).Padding(22, 18, 16).AlignItems(ui.Center).Children(func() {
		if bmp := iconBitmap(); bmp != nil {
			ui.Image(c, bmp).Size(30, 30).Radius(8).Shrink(0)
		} else {
			ui.Box(c).Size(30, 30).Radius(8).Background(ui.Hex(brandOrange)).Shrink(0)
		}
		ui.Column(c).Gap(0).Children(func() {
			ui.Text(c, "流记").Font(fontSerif).FontSize(16).Bold().SingleLine()
			ui.Text(c, "FLOWLOG").FontSize(9.5).LetterSpacing(1.4).
				TextColor(t.TextMuted).SingleLine()
		})
	})
}

// sidebarFooter 是侧边栏底部：外观切换与版本。
func (app *App) sidebarFooter(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Padding(10, 12, 12).Gap(8).Children(func() {
		ui.Box(c).Height(1).Background(t.Border).Margin(0, 4, 4)

		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			icon := iconMoon
			label := "深色"
			tip := "切换到深色"
			if app.effectiveDark(c) {
				icon, label, tip = iconSun, "浅色", "切换到浅色"
			}
			b := ui.Button(c.Key("theme-toggle"), "").Label(tip).Tooltip(tip).FontSize(11.5)
			b.Children(func() {
				ui.Icon(c, icon).FontSize(12)
				ui.Text(c, label).SingleLine()
			})
			b.OnClick(app.toggleTheme)
			ui.Spacer(c)
			ui.Text(c, "v"+appVersion).FontSize(11).TextColor(t.TextMuted)
		})
	})
}

// ---------------------------------------------------------------- 小工具

// todayEntryCount 返回今天写了几条学习记录。
func (app *App) todayEntryCount() int {
	i := app.dayIndex(app.today)
	if i < 0 {
		return 0
	}
	return len(app.days[i].Entries)
}

// openTodoCount 返回全部未完成的待办数。
func (app *App) openTodoCount() int {
	n := 0
	for _, t := range app.todos {
		if !t.Done {
			n++
		}
	}
	return n
}

// daysWritten 返回写过日记的天数。
func (app *App) daysWritten() int {
	n := 0
	for i := range app.days {
		if app.days[i].HasEntries() {
			n++
		}
	}
	return n
}

// summary 汇总全库的规模，设置页会用。
func (app *App) summary() (days, entries, points, todos int) {
	for _, d := range app.days {
		if !d.HasEntries() {
			continue
		}
		days++
		entries += len(d.Entries)
		points += d.PointCount()
	}
	return days, entries, points, len(app.todos)
}

// hasEntries 报告某天有没有日记，侧边栏日历用它打点。
func (app *App) hasEntries(date string) bool {
	i := app.dayIndex(date)
	return i >= 0 && app.days[i].HasEntries()
}

// redraw 在后台拿到新数据后请求一帧。
func (app *App) redraw() {
	if app.win != nil {
		app.win.Invalidate()
	}
}

// extractReleaseNotes 从 CHANGELOG 里取出某个版本的说明。
func extractReleaseNotes(changelog, version string) string {
	lines := strings.Split(changelog, "\n")
	start := -1
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "## ") && strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")) == version {
			start = i + 1
			break
		}
	}
	if start < 0 {
		return ""
	}
	var out []string
	for _, line := range lines[start:] {
		if strings.HasPrefix(strings.TrimSpace(line), "## ") {
			break
		}
		out = append(out, line)
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}
