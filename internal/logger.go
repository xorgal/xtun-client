package internal

import "io"

type Logger struct {
	Output io.Writer
}

func NewLogger() *Logger {
	return &Logger{}
}
