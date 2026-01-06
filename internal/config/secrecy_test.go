package config

import (
	"strings"
	"testing"
)

func TestHideStringContents(t *testing.T) {
	str := "VerySecretPassword"
	hidden := hideStringContents(str)

	if strings.Contains(hidden, str) {
		t.Errorf("hidden string contains original secret: %s", hidden)
	}

	if hidden == str {
		t.Errorf("hidden string is identical to original secret")
	}

	for i := 0; i < len(str)-3; i++ {
		substring := str[i : i+3]
		if strings.Contains(hidden, substring) {
			t.Errorf("hidden string contains original substring %q: %s", substring, hidden)
		}
	}
}
