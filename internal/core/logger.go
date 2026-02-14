package core

import (
	"os"

	"github.com/felixgeelhaar/bolt"
)

type Logger struct {
	base *bolt.Logger
}

func NewLogger(level string) *Logger {
	logger := bolt.New(bolt.NewJSONHandler(os.Stdout))
	logger.SetLevel(bolt.ParseLevel(level))
	return &Logger{base: logger}
}

func (l *Logger) Info(msg string, args ...any) {
	l.base.Info().Printf(msg, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.base.Error().Printf(msg, args...)
}
