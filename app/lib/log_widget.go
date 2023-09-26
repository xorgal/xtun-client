package lib

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/xorgal/xtun-client/internal"
)

type LogWidget struct {
	*container.Scroll
	sync.Mutex

	logs *widget.Label
}

var Log = NewLogWidget()

func NewLogWidget() *LogWidget {
	ct := internal.FormatTime(time.Now().Local())
	text := fmt.Sprintf("%s v%s log started at %s\n", internal.AppName, internal.AppVersion, ct)
	logs := widget.NewLabel(text)
	logs.Wrapping = fyne.TextWrapWord
	scroll := container.NewScroll(logs)
	return &LogWidget{scroll, sync.Mutex{}, logs}
}

func (lw *LogWidget) Write(p []byte) (n int, err error) {
	lw.Lock()
	defer lw.Unlock()

	logs := lw.logs.Text
	logs += strings.TrimSpace(string(p)) + "\n"
	lw.logs.SetText(logs)

	return len(p), nil
}

func InitAppLogger() {
	setup(Log)
}

// Setup redirects all log outputs to LogWidget.
func setup(lw *LogWidget) {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
	log.SetOutput(lw)
}
