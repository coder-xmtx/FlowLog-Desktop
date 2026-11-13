package main

import (
	"flowlog-desktop/internal/view"
	"testing"

	"github.com/egoist/mygo/ui"
)

// The view runs without a window in tests, which click and type as a user
// would.
func TestView(t *testing.T) {
	a := &view.App{}
	tt := ui.NewTester(a.View, 480, 360)
	if err := tt.Click("+"); err != nil {
		t.Fatal(err)
	}
	if a.Count != 1 || !tt.HasText("1") {
		t.Errorf("count %d after a click, texts %q", a.Count, tt.Texts())
	}
	if err := tt.Click("Name"); err != nil {
		t.Fatal(err)
	}
	tt.Type("Ada")
	if !tt.HasText("Hello, Ada!") {
		t.Errorf("texts %q after typing a name", tt.Texts())
	}
}
