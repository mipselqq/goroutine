package secrecy_test

import (
	"testing"

	"goroutine/internal/secrecy"
)

func TestSecretString(t *testing.T) {
	t.Parallel()

	const raw = "verysecret$ string123"

	secret := secrecy.SecretString(raw)

	if got := secret.RevealSecret(); got != raw {
		t.Fatalf("RevealSecret() = %q, want %q", got, raw)
	}
	if got := secret.String(); got == raw {
		t.Fatalf("String() leaked secret: %q", got)
	}
	if got := secret.LogValue().String(); got != secret.String() {
		t.Fatalf("LogValue().String() = %q, want %q", got, secret.String())
	}
}
