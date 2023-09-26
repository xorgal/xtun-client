package internal

import (
	"fmt"
	"time"
)

func FormatTime(t time.Time) string {
	d := t.Day()
	m := int(t.Month())
	y := t.Year()
	hh := t.Hour()
	mm := t.Minute()
	ss := t.Second()

	return fmt.Sprintf("%v/%v/%v %v:%v:%v", d, m, y, hh, mm, ss)
}
