package testutil

import (
	"bytes"
	"log/slog"
	"testing"
)

type testWriter struct{ t testing.TB }

func (w testWriter) Write(p []byte) (int, error) {
	s := string(p)
	s = trimTrailingNewline(s)
	w.t.Log(s)
	return len(p), nil
}

func trimTrailingNewline(s string) string {
	if s != "" && s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}

func NewLogger(t testing.TB) *slog.Logger {
	return slog.New(slog.NewTextHandler(testWriter{t}, nil))
}

func NewBufJSONLogger(t testing.TB, level slog.Level) (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: level})
	logger := slog.New(h)
	return logger, &buf
}

func NewDiscardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func FailOnInvalidLogValue(t *testing.T, attrs []slog.Attr, wantAttrs map[string]string) {
	t.Helper()

	for _, a := range attrs {
		want, ok := wantAttrs[a.Key]
		if !ok {
			t.Errorf("got unexpected attribute %q, want only configured keys", a.Key)
			continue
		}
		if a.Value.String() != want {
			t.Errorf("for key %q, got %q, want %q", a.Key, a.Value.String(), want)
		}
		delete(wantAttrs, a.Key)
	}

	for key := range wantAttrs {
		t.Errorf("missing attribute %q", key)
	}
}
