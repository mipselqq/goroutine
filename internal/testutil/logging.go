package testutil

import (
	"log/slog"
	"testing"
)

type testWriter struct{ t testing.TB }

func (w testWriter) Write(p []byte) (int, error) {
	s := string(p)
	if s != "" && s[len(s)-1] == '\n' {
		s = s[:len(s)-1]
	}
	w.t.Log(s)
	return len(p), nil
}

func CreateTestLogger(t testing.TB) *slog.Logger {
	return slog.New(slog.NewTextHandler(testWriter{t}, nil))
}
