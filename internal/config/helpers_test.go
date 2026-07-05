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

func MustNewTelegramFromEnv(t *testing.T) config.Telegram {
	t.Helper()

	cfg, err := config.NewTelegramFromEnv(testutil.NewDiscardLogger())
	if err != nil {
		t.Fatalf("NewTelegramFromEnv() error = %v", err)
	}

	return cfg
}
