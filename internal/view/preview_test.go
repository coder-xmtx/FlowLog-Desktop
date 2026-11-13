package view

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"flowlog-desktop/internal/model"
	"flowlog-desktop/internal/store"

	"github.com/egoist/mygo/ui"
)

// sampleApp 造一个带几天样本数据的 App，所有测试共用。
//
// 用 t.TempDir() 做数据目录：测试之间互不影响，也绝不碰用户真正的 data/。
func sampleApp(t *testing.T) *App {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), store.FileName))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })

	for _, d := range sampleDays() {
		if err := s.SaveDay(d); err != nil {
			t.Fatalf("SaveDay(%s): %v", d.Date, err)
		}
	}
	if err := s.SaveSettings(model.Settings{
		Theme: model.ThemeLight, UserName: "小明", Onboarded: true,
	}); err != nil {
		t.Fatal(err)
	}

	app, err := NewApp(s)
	if err != nil {
		t.Fatalf("NewApp: %v", err)
	}
	// 预览不联网：背景图留空，走渐变兜底。
	app.bg = nil
	return app
}

// sampleDays 是几天的示例数据，覆盖空备注、已勾要点、做了一半的待办等情况。
func sampleDays() []model.Day {
	today := model.MustDate(model.Today())
	return []model.Day{
		{
			Date: model.Today(),
			Entries: []model.Entry{
				{
					ID: "e-today-1", Subject: "计算机组成原理",
					Chapter: "第 4 章 存储系统", Section: "第 8 节 虚拟存储器",
					Points: []model.Point{
						{ID: "p1", Text: "页式虚拟存储器的地址映射和页表"},
						{ID: "p2", Text: "缺页中断的处理流程", Done: true},
					},
					Minutes: 45, At: "14:20",
				},
				{
					ID: "e-today-2", Subject: "数据结构",
					Chapter: "第 6 章 图", Section: "第 1 节 图的基本概念",
					Points: []model.Point{{ID: "p3", Text: "图的基本概念"}, {ID: "p4", Text: "邻接矩阵与邻接表"}},
					Note:   "先把邻接表的遍历写顺。", At: "21:05",
				},
			},
			Todos: []model.Todo{
				{ID: "t1", Text: "复习 B 树的分裂与合并", At: "09:10"},
				{ID: "t2", Text: "做一章 408 真题", Done: true, At: "09:12"},
				{ID: "t3", Text: "整理虚拟存储器的笔记", At: "14:30"},
			},
		},
		{
			Date: today.AddDate(0, 0, -1).Format(model.DateLayout),
			Entries: []model.Entry{{
				ID: "e-y1", Subject: "计算机组成原理",
				Chapter: "第 4 章 存储系统", Section: "第 8 节 虚拟存储器",
				Points: []model.Point{{ID: "p5", Text: "虚拟存储器的基本概念"}},
			}},
			Todos: []model.Todo{{ID: "t4", Text: "跑步 3 公里", Done: true, At: "18:00"}},
		},
		{
			Date: today.AddDate(0, 0, -3).Format(model.DateLayout),
			Entries: []model.Entry{
				{
					ID: "e-d3", Subject: "数据结构", Chapter: "第 6 章 树", Section: "第 3 节 二叉树的遍历",
					Points:  []model.Point{{ID: "p6", Text: "先序 / 中序 / 后序的非递归写法", Done: true}},
					Minutes: 60, At: "20:00",
				},
				{
					ID: "e-d3-2", Subject: "线性代数", Chapter: "第 5 章 特征值", Section: "第 2 节 相似对角化",
					Points: []model.Point{{ID: "p7", Text: "相似对角化的条件"}},
					Note:   "这块还不熟，下周再过一遍。", At: "22:10",
				},
			},
		},
		{
			Date: today.AddDate(0, 0, -6).Format(model.DateLayout),
			Entries: []model.Entry{{
				ID: "e-d6", Subject: "操作系统", Chapter: "第 2 章 进程", Section: "第 1 节 进程与线程",
				Points: []model.Point{{ID: "p8", Text: "进程状态转换"}, {ID: "p9", Text: "上下文切换的开销", Done: true}},
			}},
		},
	}
}

// TestPreview 渲染各个页面，把 PNG 写到 FLOWLOG_PREVIEW 指定的目录。
//
//	go test ./internal/view -run TestPreview -v
func TestPreview(t *testing.T) {
	dir := os.Getenv("FLOWLOG_PREVIEW")
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "flowlog-preview")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("建目录: %v", err)
	}

	app := sampleApp(t)

	for _, p := range []struct{ path, name string }{
		{"/", "home"},
		{"/diary", "diary"},
		{"/todo", "todo"},
		{"/settings", "settings"},
	} {
		app.router.Push(p.path)
		render(t, dir, p.name, app, 1200, 820)
	}

	// 深色
	app.settings.Theme = model.ThemeDark
	app.router.Push("/")
	render(t, dir, "home-dark", app, 1200, 820)
	app.router.Push("/diary")
	render(t, dir, "diary-dark", app, 1200, 820)
	app.router.Push("/settings")
	render(t, dir, "settings-dark", app, 1200, 820)
	app.settings.Theme = model.ThemeLight

	// 弹窗：写日记
	app.router.Push("/diary")
	app.openEditor(app.today, "")
	render(t, dir, "editor", app, 1200, 820)

	// 弹窗：编辑已有的一条
	app.ed.Open = false
	app.openEditor(model.Today(), "e-today-1")
	render(t, dir, "editor-edit", app, 1200, 820)
	app.ed.Open = false

	// 首次启动的用户名引导
	app.settings.UserName, app.settings.Onboarded = "", false
	app.onboarding = onboardState{Open: true}
	render(t, dir, "onboarding", app, 1200, 820)
	app.onboarding.Open = false
	app.settings.UserName, app.settings.Onboarded = "小明", true

	// 窄窗口
	render(t, dir, "diary-narrow", app, 900, 620)

	// 高窗口：把设置页整页展开，检查关于与更新那一节
	app.router.Push("/settings")
	render(t, dir, "settings-tall", app, 1100, 1320)

	t.Logf("预览图写到 %s", dir)
}

// render 走几帧再截图：列表要先量出行的真实高度，后面的行才摆得对位置。
func render(t *testing.T, dir, name string, app *App, w, h int) {
	t.Helper()
	tt := ui.NewTester(app.View, w, h)
	for i := 0; i < 3; i++ {
		tt.Advance(16 * time.Millisecond)
	}
	// 让提示气泡消失，别挡住截图。
	tt.Advance(5 * time.Second)

	img := tt.Image()
	path := filepath.Join(dir, name+".png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("创建 %s: %v", path, err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("写 PNG: %v", err)
	}
	t.Logf("%s (%d×%d)", path, img.Bounds().Dx(), img.Bounds().Dy())
}
