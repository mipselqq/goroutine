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

func VerifyLogValue(t *testing.T, attrs []slog.Attr, expectedAttrs map[string]string) {
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
