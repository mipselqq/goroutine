package config_test

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"goroutine/internal/config"
	"goroutine/internal/secrecy"
	"goroutine/internal/testutil"

	"github.com/google/go-cmp/cmp"
)

var defaultTelegramConfig = config.TelegramConfig{
	Token:        secrecy.SecretString("mock_token"),
	LinkTokenTTL: 15 * time.Minute,
}

var telegramEnvVars = []string{"TELEGRAM_BOT_TOKEN", "TELEGRAM_LINK_TOKEN_TTL"}

func setCustomTelegramEnvVars(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "custom_bot_token")
	t.Setenv("TELEGRAM_LINK_TOKEN_TTL", "30m")
}

func TestNewTelegramConfigFromEnv(t *testing.T) {
	t.Run("uses defaults", func(t *testing.T) {
		UnsetEnv(t, telegramEnvVars...)

		cfg := MustNewTelegramConfigFromEnv(t)

		diff := cmp.Diff(defaultTelegramConfig, cfg)
		if diff != "" {
			t.Errorf("NewTelegramConfigFromEnv() diff (-want +got):\n%s", diff)
		}
	})

	t.Run("uses env vars", func(t *testing.T) {
		setCustomTelegramEnvVars(t)

		cfg := MustNewTelegramConfigFromEnv(t)
		wantCfg := config.TelegramConfig{
			Token:        secrecy.SecretString("custom_bot_token"),
			LinkTokenTTL: 30 * time.Minute,
		}
		diff := cmp.Diff(wantCfg, cfg)
		if diff != "" {
			t.Errorf("NewTelegramConfigFromEnv() diff (-want +got):\n%s", diff)
		}
	})

	t.Run("invalid duration", func(t *testing.T) {
		t.Setenv("TELEGRAM_LINK_TOKEN_TTL", "not-a-duration")

		_, err := config.NewTelegramConfigFromEnv(testutil.NewDiscardLogger())
		if err == nil {
			t.Fatal("NewTelegramConfigFromEnv() error = nil, want non-nil")
		}
	})

	t.Run("warnings on unset variables", func(t *testing.T) {
		UnsetEnv(t, telegramEnvVars...)

		logger, buf := testutil.NewBufJsonLogger(t, slog.LevelWarn)
		_, err := config.NewTelegramConfigFromEnv(logger)
		if err != nil {
			t.Fatalf("NewTelegramConfigFromEnv() error = %v", err)
		}

		for _, envVar := range telegramEnvVars {
			if !strings.Contains(buf.String(), envVar) {
				t.Errorf("got log output %q, want mention of %q", buf.String(), envVar)
			}
		}
	})

	t.Run("no warnings if all variables are set", func(t *testing.T) {
		setCustomTelegramEnvVars(t)

		logger, buf := testutil.NewBufJsonLogger(t, slog.LevelWarn)
		_, err := config.NewTelegramConfigFromEnv(logger)
		if err != nil {
			t.Fatalf("NewTelegramConfigFromEnv() error = %v", err)
		}

		if buf.String() != "" {
			t.Errorf("got warnings %q, want none", buf.String())
		}
	})
}

func TestTelegramConfig_LogValue(t *testing.T) {
	v := defaultTelegramConfig.LogValue()
	if v.Kind() != slog.KindGroup {
		t.Fatalf("got kind %v, want Group", v.Kind())
	}

	wantAttrs := map[string]string{
		"token":          "(10 chars)",
		"link_token_ttl": "15m0s",
	}

	testutil.FailOnInvalidLogValue(t, v.Group(), wantAttrs)
}
