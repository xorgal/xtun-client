package content

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
)

func BuildDocumentationDialog(w fyne.Window) dialog.Dialog {
	text := "No documentation in this release.\n"
	d := dialog.NewInformation("Documentation", text, w)

	return d
}
