package content

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/xorgal/xtun-client/app/lib"
	"github.com/xorgal/xtun-client/internal"
	"github.com/xorgal/xtun-core/pkg/config"
	"github.com/xorgal/xtun-core/pkg/counter"
)

type HomeScreen struct {
	w               fyne.Window
	addrLabel       *widget.Label
	ctrlBtn         *widget.Button
	statsForm       *widget.Form
	serverIPLabel   *widget.Label
	clientIPLabel   *widget.Label
	bufferSizeLabel *widget.Label
	mtuLabel        *widget.Label
	compressLabel   *widget.Label
	readBytes       *widget.Label
	writeBytes      *widget.Label
	errCh           chan<- error
	container       *fyne.Container
}

func BuildHomeScreen(w fyne.Window) *fyne.Container {
	var s HomeScreen

	s.w = w

	s.addrLabel = widget.NewLabel(config.AppConfig.ServerAddr)
	s.addrLabel.Alignment = fyne.TextAlignCenter
	s.addrLabel.TextStyle = fyne.TextStyle{Bold: true}

	s.ctrlBtn = s.buildCtrlBtn()

	s.serverIPLabel = widget.NewLabel(config.AppConfig.ServerIP)
	s.clientIPLabel = widget.NewLabel(strings.Split(config.AppConfig.CIDR, "/")[0])
	s.bufferSizeLabel = widget.NewLabel(strconv.Itoa(config.AppConfig.BufferSize))
	s.mtuLabel = widget.NewLabel(strconv.Itoa(config.AppConfig.MTU))
	s.compressLabel = widget.NewLabel(strconv.FormatBool(config.AppConfig.Compress))
	s.readBytes = widget.NewLabel("")
	s.writeBytes = widget.NewLabel("")

	// Initialize statsForm with the read and write labels
	s.statsForm = widget.NewForm(
		widget.NewFormItem("Server IP", s.serverIPLabel),
		widget.NewFormItem("Client IP", s.clientIPLabel),
		widget.NewFormItem("Buffer Size", s.bufferSizeLabel),
		widget.NewFormItem("MTU", s.mtuLabel),
		widget.NewFormItem("Use Compression", s.compressLabel),
		widget.NewFormItem("Read Bytes", s.readBytes),
		widget.NewFormItem("Written Bytes", s.writeBytes),
	)
	s.statsForm.Hide()

	s.errCh = make(chan<- error)

	s.container = container.NewVBox(s.addrLabel, s.ctrlBtn, s.statsForm)

	state := getConnectionStateNotifier()

	go s.updateScreen(state)

	return s.container
}

func (s *HomeScreen) updateScreen(state chan internal.ConnectionState) {
	for state := range state {
		switch state {
		case internal.Disconnected:
			s.ctrlBtn.Text = "Connect"
			s.ctrlBtn.OnTapped = s.connect
			s.ctrlBtn.Enable()
			s.statsForm.Hide() // Hide stats when disconnected
		case internal.Connected:
			s.ctrlBtn.Text = "Disconnect"
			s.ctrlBtn.OnTapped = s.disconnect
			s.ctrlBtn.Enable()
			s.statsForm.Show() // Show stats when connected
			// Update read and write labels
			s.readBytes.SetText(formatBytes(counter.GetReadBytes()))
			s.writeBytes.SetText(formatBytes(counter.GetWrittenBytes()))
		case internal.Connecting:
			s.ctrlBtn.Text = "Connecting..."
			s.ctrlBtn.Disable()
			s.statsForm.Hide() // Hide stats when connecting
		case internal.Disconnecting:
			s.ctrlBtn.Text = "Disconnecting..."
			s.ctrlBtn.Disable()
			s.statsForm.Hide() // Hide stats when disconnecting
		}

		s.container.Refresh()
	}
}

func (s *HomeScreen) buildCtrlBtn() *widget.Button {
	var label string
	var action func()

	state := internal.GetConnectionState()

	if state == internal.Connected {
		label = "Disconnect"
		action = s.disconnect
	} else {
		label = "Connect"
		action = s.connect
	}

	return widget.NewButton(label, action)
}

func (s *HomeScreen) connect() {
	go internal.StartClient(config.AppConfig, s.errCh)
}

func (s *HomeScreen) disconnect() {
	err := internal.StopClient(config.AppConfig)
	if err != nil {
		lib.ShowErrorDialog(s.w, err)
		log.Print(err)
	}
}

func getConnectionStateNotifier() chan internal.ConnectionState {
	state := make(chan internal.ConnectionState)

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for range ticker.C {
			current := internal.GetConnectionState()
			state <- current
		}
	}()

	counter.GetReadBytes()
	return state
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d bytes", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cb", float64(b)/float64(div), "KMGTPE"[exp])
}
