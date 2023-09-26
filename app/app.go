package app

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"github.com/xorgal/xtun-client/app/content"
	"github.com/xorgal/xtun-client/internal"
)

func RunLoop() {
	a := app.New()
	w := a.NewWindow("xtun client")

	// Build main menu
	mainMenu := BuildMainMenu(a, w)

	var currentScreen fyne.CanvasObject
	if internal.AppState.IsInitialized {
		currentScreen = content.BuildHomeScreen(w)
	} else {
		currentScreen = content.BuildSetupScreen(w)
	}

	// Setup app in system's tray
	SetSystemTray(a, w)

	// Set size and apply window properties
	SetWindowProps(w)

	// Set main menu
	w.SetMainMenu(mainMenu)

	// Set content
	w.SetContent(currentScreen)

	// Start RunLoop
	w.ShowAndRun()
}
