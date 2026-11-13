package view

import (
	"fmt"

	"github.com/egoist/mygo/ui"
)

// app is the state the window shows. Its view builds the interface from
// it, on the main thread, whenever the window needs a frame: after input,
// after Window.Update, and while something animates.
type App struct {
	Name  string
	Count int
}

func (a *App) View(c *ui.Context) {
	t := c.Theme()
	ui.Column(c).Fill().Center().Gap(16).Padding(24).Children(func() {
		ui.Text(c, "FlowLog Desktop").FontSize(26).Bold()
		ui.TextInput(c, &a.Name).Placeholder("Your name").Label("Name").Width(240)
		greeting := "Hello!"
		if a.Name != "" {
			greeting = fmt.Sprintf("Hello, %s!", a.Name)
		}
		ui.Text(c, greeting).TextColor(t.TextMuted)
		ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
			if ui.Button(c, "−").Width(44).Clicked() {
				a.Count--
			}
			ui.Textf(c, "%d", a.Count).FontSize(20).Width(48).TextAlign(ui.Center)
			if ui.PrimaryButton(c, "+").Width(44).Clicked() {
				a.Count++
			}
		})
	})
}
