package config_test

import (
	"testing"

	"goroutine/internal/config"
	"goroutine/internal/testutil"
)

func UnsetEnv(t *testing.T, keys ...string) {
	t.Helper()

	for _, key := range keys {
		t.Setenv(key, "")
	}
}

func MustNewTelegramConfigFromEnv(t *testing.T) config.TelegramConfig {
	t.Helper()

	cfg, err := config.NewTelegramConfigFromEnv(testutil.NewDiscardLogger())
	if err != nil {
		t.Fatalf("NewTelegramConfigFromEnv() error = %v", err)
	}

	return cfg
}
