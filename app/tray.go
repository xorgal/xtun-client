package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

func SetSystemTray(a fyne.App, w fyne.Window) {
	if desk, ok := a.(desktop.App); ok {
		m := fyne.NewMenu("xtun client",
			fyne.NewMenuItem("Open app", func() {
				w.Show()
			}),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() {
				a.Quit()
			}),
		)
		m.Label = "xtun"

		desk.SetSystemTrayMenu(m)
	}

	w.SetContent(widget.NewLabel("System Tray"))

	w.SetCloseIntercept(func() {
		w.Hide()
	})
}
