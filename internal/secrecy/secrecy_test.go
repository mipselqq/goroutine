package secrecy_test

import (
	"encoding/json"
	"strings"
	"testing"

	"goroutine/internal/secrecy"
)

func TestSecretString(t *testing.T) {
	t.Parallel()

	const raw = "verysecret$ string123"

	secret := secrecy.SecretString(raw)

	got := secret.RevealSecret()
	if got != raw {
		t.Fatalf("RevealSecret() = %q, want %q", got, raw)
	}
	if got = secret.String(); got == raw {
		t.Fatalf("String() leaked secret: %q", got)
	}
	got = secret.LogValue().String()
	if got != secret.String() {
		t.Fatalf("LogValue().String() = %q, want %q", got, secret.String())
	}

	marshaled, err := json.Marshal(secret)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	got = string(marshaled)
	if strings.Contains(got, raw) {
		t.Fatalf("json.Marshal() leaked secret: %q", got)
	}
}
