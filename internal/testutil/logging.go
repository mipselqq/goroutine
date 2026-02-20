package testutil

import (
	"bytes"
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

func NewTestLogger(t testing.TB) *slog.Logger {
	return slog.New(slog.NewTextHandler(testWriter{t}, nil))
}

func NewBufJsonLogger(t testing.TB) (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	h := slog.NewJSONHandler(&buf, nil)
	logger := slog.New(h)
	return logger, &buf
}

func NewDiscardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func FailOnInvalidLogValue(t *testing.T, attrs []slog.Attr, expectedAttrs map[string]string) {
	t.Helper()

	for _, a := range attrs {
		expected, ok := expectedAttrs[a.Key]
		if !ok {
			t.Errorf("unexpected attribute %q", a.Key)
			continue
		}
		if a.Value.String() != expected {
			t.Errorf("for key %q, expected %q, got %q", a.Key, expected, a.Value.String())
		}
		delete(expectedAttrs, a.Key)
	}

	for key := range expectedAttrs {
		t.Errorf("missing attribute %q", key)
	}
}
