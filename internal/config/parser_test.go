package config

import (
	"testing"
)

func TestGetenvOrDefault(t *testing.T) {
	key := "TEST_ENV_VAR"
	val := "some_value"
	def := "default_value"

	t.Run("returns value from env", func(t *testing.T) {
		t.Setenv(key, val)
		if got := getenvOrDefault(key, def); got != val {
			t.Errorf("expected %q, got %q", val, got)
		}
	})

	t.Run("returns default value if not set", func(t *testing.T) {
		if got := getenvOrDefault("NON_EXISTENT_VAR_XYZ_123", def); got != def {
			t.Errorf("expected %q, got %q", def, got)
		}
	})

	t.Run("returns default value if empty", func(t *testing.T) {
		t.Setenv(key, "")
		if got := getenvOrDefault(key, def); got != def {
			t.Errorf("expected %q, got %q", def, got)
		}
	})
}
