package content

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"github.com/xorgal/xtun-client/internal"
)

func BuildAboutDialog(w fyne.Window) dialog.Dialog {
	l1 := "Client application to establish connection\n"
	l2 := "with xtun Virtual Private Network (VPN)\n"
	l3 := "over WebSocket protocol.\n\n"
	l4 := fmt.Sprintf("Version: %v.\n\n", internal.AppVersion)
	l5 := "(c) 2023 All rights reserved."
	text := fmt.Sprintf("%s%s%s%s%s\n", l1, l2, l3, l4, l5)
	d := dialog.NewInformation(internal.AppName, text, w)

	return d
}
