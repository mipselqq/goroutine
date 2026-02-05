package secrecy_test

import (
	"strings"
	"testing"

	"go-todo/internal/secrecy"
)

func TestSecretString(t *testing.T) {
	s := "verysecret$ string123"

	ss := secrecy.SecretString(s)
	ssRevealed := ss.RevealSecret()

	if ss.RevealSecret() != s {
		t.Errorf("Expected '%s' after reveal, got '%s'", s, ssRevealed)
	}

	ssHiddenRepr := ss.String()
	ssLogged := ss.LogValue().String()

	if ssLogged != ssHiddenRepr {
		t.Errorf("Expected LogValue '%s' == String '%s'", ssLogged, ssHiddenRepr)
	}

	sLower := strings.ToLower(s)
	ssLower := strings.ToLower(ssHiddenRepr)

	mid := len(sLower) / 2
	left := sLower[mid:]
	right := sLower[:mid]

	if strings.Contains(ssLower, left) || strings.Contains(ssLower, right) {
		t.Errorf("Secret representation '%s' contains half of original '%s'", ss, s)
	}
}
