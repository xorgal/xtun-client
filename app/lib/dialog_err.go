package lib

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func ShowErrorDialog(w fyne.Window, err error) {
	log.Print(err)
	dialog.NewError(err, w).Show()
}
