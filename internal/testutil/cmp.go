package testutil

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"goroutine/internal/domain"
)

type SecretValue interface {
	fmt.Stringer
	LogValue() slog.Value
	MarshalJSON() ([]byte, error)
}

func AssertSecretHidden(t *testing.T, raw string, secret SecretValue) {
	t.Helper()

	rawLower := strings.ToLower(raw)

	got := secret.String()
	if strings.Contains(strings.ToLower(got), rawLower) {
		t.Fatalf("String leaked raw secret: got %q, raw %q", got, raw)
	}

	verbs := []string{"%v", "%s", "%q", "%+v", "%#v"}
	for _, verb := range verbs {
		got = fmt.Sprintf(verb, secret)
		if strings.Contains(strings.ToLower(got), rawLower) {
			t.Fatalf("verb %s leaked raw secret: got %q, raw %q", verb, got, raw)
		}
	}

	got = secret.LogValue().String()
	if strings.Contains(strings.ToLower(got), rawLower) {
		t.Fatalf("LogValue().String() leaked raw secret: got %q, raw %q", got, raw)
	}

	marshaled, err := json.Marshal(secret)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	got = string(marshaled)
	if strings.Contains(strings.ToLower(got), rawLower) {
		t.Fatalf("json.Marshal() leaked raw secret: got %q, raw %q", got, raw)
	}
}

func CmpAllowUnexported() cmp.Option {
	return cmp.AllowUnexported(
		domain.BoardID{},
		domain.BoardName{},
		domain.BoardDescription{},
		domain.UserID{},
		domain.ColumnID{},
		domain.ColumnName{},
		domain.ColumnDescription{},
		domain.ColumnPosition{},
		domain.TaskID{},
		domain.TaskName{},
		domain.TaskDescription{},
		domain.TaskPosition{},
		domain.UserPassword{},
		domain.AuthToken{},
	)
}
