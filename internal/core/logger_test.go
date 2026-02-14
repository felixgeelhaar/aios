package core

import "testing"

func TestLoggerInfo(t *testing.T) {
	l := NewLogger("info")
	// Should not panic.
	l.Info("test message %s", "arg")
}

func TestLoggerError(t *testing.T) {
	l := NewLogger("error")
	// Should not panic.
	l.Error("test error %s", "arg")
}

func TestLoggerDebugLevel(t *testing.T) {
	l := NewLogger("debug")
	l.Info("debug level info %d", 42)
	l.Error("debug level error %d", 42)
}
