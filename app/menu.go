package app

import (
	"fyne.io/fyne/v2"
	"github.com/xorgal/xtun-client/app/content"
	"github.com/xorgal/xtun-client/app/lib"
)

func BuildMainMenu(a fyne.App, w fyne.Window) *fyne.MainMenu {
	mainMenu := fyne.NewMainMenu(
		fyne.NewMenu("File",
			fyne.NewMenuItem("Home", func() { w.SetContent(content.BuildHomeScreen(w)) }),
			fyne.NewMenuItem("Log", func() { w.SetContent(lib.Log) }),
			fyne.NewMenuItem("Preferenses", func() { w.SetContent(content.BuildSetupScreen(w)) }),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("Quit", func() { a.Quit() }),
		),
		fyne.NewMenu("Help",
			fyne.NewMenuItem("Documentation", func() { content.BuildDocumentationDialog(w).Show() }),
			fyne.NewMenuItemSeparator(),
			fyne.NewMenuItem("About", func() { content.BuildAboutDialog(w).Show() }),
		),
	)

	return mainMenu
}
