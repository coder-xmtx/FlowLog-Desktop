package app

import (
	"path/filepath"
	"testing"

	"flowlog-desktop/internal/model"
	"flowlog-desktop/internal/store"
)

// newTestApp 返回一个空的 App（只有今天这一天）与它的数据库。
func newTestApp(t *testing.T) *App {
	t.Helper()
	s, err := store.Open(filepath.Join(t.TempDir(), store.FileName))
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	st := model.DefaultSettings()
	st.Onboarded = true
	if err := s.SaveSettings(st); err != nil {
		t.Fatal(err)
	}
	a, err := New(s)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return a
}

// addEntry 用编辑器写入一条记录，返回它的 ID。
func addEntry(t *testing.T, a *App, subject, points string) string {
	t.Helper()
	a.OpenEditor(a.Today, "")
	a.Editor.Subject = subject
	a.Editor.Points = points
	a.SaveEditor()
	if len(a.Day(a.Today).Entries) == 0 {
		t.Fatalf("没有写入记录，Err=%q", a.Editor.Err)
	}
	day := a.Day(a.Today)
	return day.Entries[len(day.Entries)-1].ID
}

// TestSaveEditor 检查校验与落库：没科目、没要点都不该存。
func TestSaveEditor(t *testing.T) {
	a := newTestApp(t)

	a.OpenEditor(a.Today, "")
	a.SaveEditor()
	if a.Editor.Err == "" || !a.Editor.Open {
		t.Fatalf("空表单应该校验失败并保持打开")
	}

	a.Editor.Subject = "数据结构"
	a.SaveEditor()
	if a.Editor.Err == "" {
		t.Fatalf("没有要点和备注时应该校验失败")
	}

	a.Editor.Points = "- 图的基本概念\n邻接矩阵"
	a.SaveEditor()
	if a.Editor.Open || a.Editor.Err != "" {
		t.Fatalf("有效内容应该保存成功: %+v", a.Editor)
	}
	day := a.Day(a.Today)
	if len(day.Entries) != 1 || len(day.Entries[0].Points) != 2 {
		t.Fatalf("落库结果 = %+v", day.Entries)
	}
	if day.Entries[0].Points[0].Text != "图的基本概念" {
		t.Errorf("开头的「- 」应该被去掉: %q", day.Entries[0].Points[0].Text)
	}

	// 重新打开数据库确认真的写进去了。
	reopened, err := store.Open(a.Store.Path())
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	days, _ := reopened.Days()
	if len(days) != 1 || len(days[0].Entries[0].Points) != 2 {
		t.Fatalf("数据库里 = %+v", days)
	}
}

// TestSubjectFilterAndSearch 检查科目筛选与搜索的叠加。
func TestSubjectFilterAndSearch(t *testing.T) {
	a := newTestApp(t)
	addEntry(t, a, "数据结构", "图的基本概念\n邻接矩阵")
	addEntry(t, a, "计算机组成原理", "虚拟存储器")

	if got := len(a.FilterDays()); got != 1 {
		t.Fatalf("今天应该有 1 天, 得到 %d", got)
	}

	a.SubjectFilter = "数据结构"
	shown := a.ShownEntries(a.Day(a.Today))
	if len(shown) != 1 || a.Day(a.Today).Entries[shown[0]].Subject != "数据结构" {
		t.Fatalf("科目筛选后应只剩数据结构: %+v", shown)
	}

	// 搜索词和科目叠加：换个科目就不该命中。
	a.Query = "邻接矩阵"
	if got := len(a.ShownEntries(a.Day(a.Today))); got != 1 {
		t.Errorf("数据结构 + 邻接矩阵 应命中 1 条, 得到 %d", got)
	}
	a.SubjectFilter = "计算机组成原理"
	if got := len(a.FilterDays()); got != 0 {
		t.Errorf("计组 + 邻接矩阵 不该命中, 得到 %d 天", got)
	}
}

// TestDeleteAndUndo 检查删除一条记录后能撤销回来。
func TestDeleteAndUndo(t *testing.T) {
	a := newTestApp(t)
	id := addEntry(t, a, "数据结构", "图的基本概念")

	a.DeleteEntry(a.Today, id)
	if len(a.Day(a.Today).Entries) != 0 {
		t.Fatalf("删除后还有记录")
	}

	// 取走提示里的撤销动作并执行。
	var undo func()
	for _, toast := range a.TakeToasts() {
		if toast.OnAct != nil {
			undo = toast.OnAct
		}
	}
	if undo == nil {
		t.Fatalf("没有可撤销的提示")
	}
	undo()
	if len(a.Day(a.Today).Entries) != 1 {
		t.Fatalf("撤销后应该恢复 1 条")
	}
}

// TestTodoFlow 检查待办的添加、勾选、删除与撤销。
func TestTodoFlow(t *testing.T) {
	a := newTestApp(t)

	a.TodoDraft = "复习 B 树"
	a.SetDue(1) // 明天
	a.AddTodo()
	if len(a.Todos) != 1 {
		t.Fatalf("待办数 = %d", len(a.Todos))
	}
	todo := a.Todos[0]
	if todo.Text != "复习 B 树" || todo.Due != model.DayAfter(a.Today, 1) {
		t.Fatalf("新待办 = %+v", todo)
	}
	if a.TodoDraft != "" || !a.TodoDue.IsZero() {
		t.Errorf("添加后输入框与截止日期都应清空")
	}

	todo.Done = true
	a.SaveTodo(todo)
	if !a.Todos[0].Done {
		t.Errorf("勾选没有写回")
	}

	a.DeleteTodo(todo.ID)
	if len(a.Todos) != 0 {
		t.Fatalf("删除后还有待办")
	}
	toasts := a.TakeToasts()
	if len(toasts) == 0 || toasts[len(toasts)-1].OnAct == nil {
		t.Fatalf("删除后应该有撤销提示")
	}
	toasts[len(toasts)-1].OnAct()
	if len(a.Todos) != 1 || !a.Todos[0].Done {
		t.Errorf("撤销后 = %+v", a.Todos)
	}
}

// TestTodoRowsFilter 检查「进行中 / 已完成」的筛选。
func TestTodoRowsFilter(t *testing.T) {
	a := newTestApp(t)
	a.Todos = []model.Todo{{ID: "a", Text: "没做的"}, {ID: "b", Text: "做完的", Done: true}}

	a.TodoFilter = 0
	if got := len(a.TodoRows()); got != 2 {
		t.Errorf("全部 = %d", got)
	}
	a.TodoFilter = 1
	if rows := a.TodoRows(); len(rows) != 1 || a.Todos[rows[0]].ID != "a" {
		t.Errorf("进行中 = %v", rows)
	}
	a.TodoFilter = 2
	if rows := a.TodoRows(); len(rows) != 1 || a.Todos[rows[0]].ID != "b" {
		t.Errorf("已完成 = %v", rows)
	}
}

// TestAutoBackup 检查启动备份：每天一份，可在设置里关掉。
func TestAutoBackup(t *testing.T) {
	a := newTestApp(t)
	addEntry(t, a, "数据结构", "图")

	a.AutoBackup()
	info, ok := store.LatestBackup(a.BackupDir())
	if !ok || info.Name == "" {
		t.Fatalf("没有生成备份")
	}

	// 再跑一次：今天已经备份过，不该多出一份。
	a.AutoBackup()
	again, _ := store.LatestBackup(a.BackupDir())
	if again.Name != info.Name {
		t.Errorf("同一天备份了两次: %q → %q", info.Name, again.Name)
	}

	// 关掉开关就不再备份。
	a.Settings.AutoBackup = false
	a.AutoBackup()
	last, _ := store.LatestBackup(a.BackupDir())
	if last.Name != info.Name {
		t.Errorf("关掉开关后仍然备份了")
	}
}

// TestSettingsAndOnboarding 检查称呼引导与外观设置的落库。
func TestSettingsAndOnboarding(t *testing.T) {
	a := newTestApp(t)
	a.Onboarding = OnboardingState{Open: true}

	a.FinishOnboarding()
	if a.Onboarding.Err == "" {
		t.Errorf("空称呼应该提示错误")
	}

	a.Onboarding.Name = "小明"
	a.FinishOnboarding()
	if a.Settings.UserName != "小明" || !a.Settings.Onboarded {
		t.Errorf("设置 = %+v", a.Settings)
	}

	a.SetTheme(model.ThemeLight)
	st, err := a.Store.Settings()
	if err != nil {
		t.Fatal(err)
	}
	if st.Theme != model.ThemeLight {
		t.Errorf("外观没有落库: %+v", st)
	}
}

// TestStreak 检查连续记录：今天没写不算断。
func TestStreak(t *testing.T) {
	a := newTestApp(t)
	today := model.MustDate(a.Today)
	for i := 1; i <= 2; i++ {
		d := today.AddDate(0, 0, -i).Format(model.DateLayout)
		a.Upsert(model.Day{Date: d, Entries: []model.Entry{{ID: d, Subject: "x"}}})
	}
	if got := a.Streak(); got != 2 {
		t.Errorf("streak = %d, 期望 2", got)
	}
}
