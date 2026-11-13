// Package view 是流记的界面层：状态、路由、四个页面与各种弹层。
//
// 这一层遵守 MyGO 原生 UI 的三条硬规则：
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
type App struct {
	store  *store.Store
	router *ui.Router
	win    *mygo.Window

	// days 是所有有内容的日子，按日期倒序；今天永远在列表里（哪怕还空着）。
	days  []model.Day
	today string

	// 外观。用下标而不是字符串，方便直接绑到 Segmented。
	themePick  int
	accentPick int

	// 日记页
	diaryList   ui.ListState
	query       string
	focusSearch bool
	filtered    []int // 每帧按搜索词算出的 app.days 下标

	// 待办页
	todoList   ui.ListState
	todoDraft  string
	todoFilter int
	todoRows   []int // 每帧按筛选算出的「今天待办」下标

	// 弹层
	ed         editorState
	confirm    confirmState
	appearance bool

	// 待展示的提示，在下一帧开头交给 c.AddToast。
	toasts []pendingToast
}

// editorState 是「写日记」弹窗的草稿。
//
// 它是一份普通状态：弹窗打开时从 Day 里读出来，关闭时丢弃，
// 所以用户点到外面把弹窗关掉也不会污染数据。
type editorState struct {
	Open    bool
	Date    time.Time
	EntryID string // 正在编辑的条目 ID；空字符串表示新增
	Subject string
	Topic   string
	Points  string // 一行一条，支持从旧 Markdown 直接粘贴
	Note    string
	Minutes float64
	Err     string
}

// confirmState 是一个「确认一下」的弹窗。
type confirmState struct {
	Open    bool
	Title   string
	Message string
	Yes     string
	YesFunc func()
}

// pendingToast 是一条等到下一帧再交给 MyGO 的提示。
type pendingToast struct {
	Title  string
	Desc   string
	Type   string
	Action string
	OnAct  func()
}

// NewApp 载入数据目录，建立视图状态。
//
// 读不出来的文件不会拦住启动：坏文件记一条提示，其余照常显示。
func NewApp(s *store.Store) *App {
	app := &App{
		store:  s,
		router: ui.NewRouter("/"),
		today:  model.Today(),
	}

	days, bad, err := s.LoadDays()
	switch {
	case err != nil:
		app.toastError("读不了数据目录", err.Error())
	case len(bad) > 0:
		app.toastError("有文件读不出来", "已跳过："+strings.Join(bad, "、"))
	}
	app.days = days

	st, err := s.LoadSettings()
	if err != nil {
		app.toastError("读不了设置", err.Error())
	}
	app.themePick = themeModeIndex(st.Theme)
	app.accentPick = accentIndexOf(st.Accent)

	app.ensureToday()
	return app
}

// SetWindow 记住主窗口，供菜单项之类的入口在需要时请求重绘。
func (app *App) SetWindow(w *mygo.Window) { app.win = w }

// NewEntry 打开「写今天的日记」，供应用菜单调用。
//
// 菜单回调不在视图构建里，改完状态要自己请求重绘：
// Update 会在主线程上重建一帧。
func (app *App) NewEntry() {
	app.openEditor(app.today, "")
	if app.win != nil {
		app.win.Invalidate()
	}
}

// ShowPage 跳到某个页面，供应用菜单调用。
func (app *App) ShowPage(path string) {
	app.router.Push(path)
	if app.win != nil {
		app.win.Invalidate()
	}
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
			case r.Match("/about"):
				r.Title("流记 · 关于")
				app.about(c)
			default:
				ui.Text(c, "没有这个页面")
			}
		})
	})

	// 弹层写在路由外面：切换页面时它们不会跟着消失，
	// 而且 Modal / AlertDialog 本身就画在窗口内容之上。
	app.diaryEditor(c)
	app.confirmDialog(c)
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

// day 返回某一天的数据；不存在时返回一个空 Day（不写盘）。
func (app *App) day(date string) model.Day {
	if i := app.dayIndex(date); i >= 0 {
		return app.days[i]
	}
	return model.Day{Date: date, Entries: []model.Entry{}, Todos: []model.Todo{}}
}

// ensureToday 保证今天在列表里，这样「今天」永远有一张可以落笔的卡片。
func (app *App) ensureToday() {
	if app.dayIndex(app.today) >= 0 {
		return
	}
	pos := sort.Search(len(app.days), func(i int) bool { return app.days[i].Date <= app.today })
	app.days = append(app.days, model.Day{})
	copy(app.days[pos+1:], app.days[pos:])
	app.days[pos] = model.Day{Date: app.today, Entries: []model.Entry{}, Todos: []model.Todo{}}
}

// rollOver 处理跨零点：应用一直开着的时候，「今天」要跟着日期走，
// 否则第二天用户会继续写到昨天那一条上。
func (app *App) rollOver() {
	today := model.Today()
	if today == app.today {
		return
	}
	app.today = today
	app.ensureToday()
}

// persist 把某一天的当前内容写盘。
//
// 调用点都在 OnClick / OnChange 里，也就是本帧构建完成之后，
// 所以这里改 app.days 不会影响正在遍历的列表。
func (app *App) persist(date string) {
	i := app.dayIndex(date)
	if i < 0 {
		return
	}
	if err := app.store.SaveDay(app.days[i]); err != nil {
		app.toastError("没有保存成功", err.Error())
		return
	}
	app.pruneEmpty()
}

// pruneEmpty 丢掉「已经不是今天、且一条内容都没有」的日子。
//
// 对应 store 的行为：空的一天不留 JSON 文件，界面上也不该留一张空卡片。
func (app *App) pruneEmpty() {
	kept := app.days[:0]
	for _, d := range app.days {
		if d.HasContent() || d.Date == app.today {
			kept = append(kept, d)
		}
	}
	app.days = kept
}

// saveSettings 把外观写进 data/settings.json。
func (app *App) saveSettings() {
	st := model.Settings{Theme: themeModeAt(app.themePick), Accent: accentChoices[app.accentPick].Hex}
	if err := app.store.SaveSettings(st); err != nil {
		app.toastError("没有保存外观设置", err.Error())
	}
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
			ID:          t.Title, // 同一个标题的提示会就地更新，而不是叠一摞
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

// shortcuts 处理窗口级快捷键。
//
// 弹窗开着时 c.Shortcut 自动失效（焦点被弹窗接管），所以不用额外判断。
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
		app.router.Push("/about")
	case c.Shortcut(ui.Cmd, ui.KeyF):
		app.router.Push("/diary")
		app.focusSearch = true
	}
}

// ---------------------------------------------------------------- 侧边栏

// sidebar 是左侧导航：品牌、四个页面、外观入口。
//
// 用 ui.Sidebar 而不是自己拼按钮，是为了白拿键盘导航、选中态和
// 无障碍树（读屏会把它读成一个源列表）。
func (app *App) sidebar(c *ui.Context) {
	t := c.Theme()
	side := ui.Column(c).Width(236).Shrink(0).FillHeight().
		Background(t.Surface).BorderWidth(0, 1, 0, 0).BorderColor(t.Border)

	side.Children(func() {
		app.brand(c)

		page := app.router.Path()
		if ui.Sidebar(c, &page, func() {
			ui.SidebarItem(c, "/", iconHome, "首页")
			item := ui.SidebarItem(c, "/diary", iconBook, "日记")
			if n := app.daysWritten(); n > 0 {
				item.Children(func() { ui.Badge(c, strconv.Itoa(n)) })
			}
			todo := ui.SidebarItem(c, "/todo", iconCheckCircle, "待办")
			if open := app.openTodoCount(); open > 0 {
				todo.Children(func() { ui.Badge(c, strconv.Itoa(open)) })
			}
			ui.SidebarItem(c, "/about", iconInfo, "关于")
		}).Grow(1).Label("主导航").Changed() {
			app.router.Push(page)
		}

		app.sidebarFooter(c)
	})
}

// brand 是侧边栏顶部的标志。
func (app *App) brand(c *ui.Context) {
	t := c.Theme()
	ui.Row(c).Gap(10).Padding(22, 18, 16).AlignItems(ui.Center).Children(func() {
		// logo 用图标上的亮橙，界面强调色则是更深的同色系。
		logo := ui.Box(c).Size(30, 30).Radius(9).Background(ui.Hex(brandOrange)).Shrink(0)
		logo.Draw(func(p *ui.Painter, r ui.Rect) {
			drawLogoMark(p, r, ui.RGBA(255, 255, 255, 0.96), 2)
		})
		ui.Column(c).Gap(0).Children(func() {
			ui.Text(c, "流记").FontSize(15).Bold().SingleLine()
			ui.Text(c, "FLOWLOG").FontSize(9.5).LetterSpacing(1.4).
				TextColor(t.TextMuted).SingleLine()
		})
	})
}

// sidebarFooter 是侧边栏底部：外观入口与版本。
func (app *App) sidebarFooter(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Padding(12, 12, 14).Gap(10).Children(func() {
		ui.Box(c).Height(1).Background(t.Border).Margin(0, 4, 4)

		btn := ui.Button(c, "外观 · "+themeModeName(themeModeAt(app.themePick))).
			FontSize(12).Justify(ui.Start).FillWidth().Label("外观设置").
			Tooltip("切换浅色 / 深色 / 强调色")
		btn.Children(func() {
			ui.Icon(c, app.themeIcon()).FontSize(14)
		})
		// Popover 只负责「开着的时候画在哪」，开关要自己管。
		btn.OnClick(func() { app.appearance = !app.appearance })

		ui.Popover(c, btn, &app.appearance, func() {
			ui.Column(c).Width(232).Gap(10).Padding(4).Children(func() {
				ui.Text(c, "外观").FontSize(11.5).FontWeight(600).TextColor(t.TextMuted)
				ui.Segmented(c, &app.themePick, "跟随", "浅色", "深色").
					FillWidth().OnChange(app.saveSettings)

				ui.Box(c).Height(1).Background(t.Border).Margin(4, 0)
				ui.Text(c, "强调色").FontSize(11.5).FontWeight(600).TextColor(t.TextMuted)
				ui.Row(c).Gap(9).Wrap().Children(func() {
					for i, a := range accentChoices {
						app.accentSwatch(c, i, a)
					}
				})
			})
		})

		ui.Text(c, "v"+appVersion+" · MyGo 原生 UI").FontSize(10.5).TextColor(t.TextMuted)
	})
}

// themeIcon 返回当前外观对应的图标。
func (app *App) themeIcon() *ui.SVG {
	switch themeModeAt(app.themePick) {
	case model.ThemeLight:
		return iconSun
	case model.ThemeDark:
		return iconMoon
	}
	return iconMonitor
}

// accentSwatch 是一个可选中的强调色圆点。
func (app *App) accentSwatch(c *ui.Context, i int, a accentChoice) {
	t := c.Theme()
	box := ui.Box(c.Key("accent-"+a.Hex)).Size(22, 22).Radius(11).
		Background(ui.Hex(a.Hex)).Cursor(ui.CursorPointer).
		Label(a.Name).Tooltip(a.Name)
	if i == app.accentPick {
		box.Border(2, t.Text).Padding(0)
	} else {
		box.Border(1, ui.RGBA(0, 0, 0, 0.12))
	}
	box.OnClick(func() {
		app.accentPick = i
		app.saveSettings()
	})
}

// ---------------------------------------------------------------- 小工具

// daysWritten 返回真正写过东西的天数。
//
// 只数不切片：它在侧边栏里每帧都要用，别为了一个数字每帧分配一个切片。
func (app *App) daysWritten() int {
	n := 0
	for i := range app.days {
		if app.days[i].HasContent() {
			n++
		}
	}
	return n
}

// openTodoCount 返回今天还没做的待办数。
func (app *App) openTodoCount() int {
	done, total := app.day(app.today).TodoProgress()
	return total - done
}
