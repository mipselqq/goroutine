package domain_test

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
)

type secretValue interface {
	fmt.Stringer
	LogValue() slog.Value
}

func assertSecretHidden(t *testing.T, raw string, secret secretValue) {
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
