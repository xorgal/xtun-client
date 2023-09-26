package app

import "fyne.io/fyne/v2"

func SetWindowProps(w fyne.Window) {
	w.SetFixedSize(true)
	w.Resize(fyne.NewSize(450, 500))
}
