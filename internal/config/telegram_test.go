package config_test

import (
	"log/slog"
	"strings"
	"testing"

	"goroutine/internal/config"
	"goroutine/internal/secrecy"
	"goroutine/internal/testutil"

	"github.com/google/go-cmp/cmp"
)

var defaultTelegramConfig = config.TelegramConfig{
	Token: secrecy.SecretString("mock_token"),
}

var telegramEnvVars = []string{"TELEGRAM_BOT_TOKEN"}

func setCustomTelegramEnvVars(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", "custom_bot_token")
}

func TestNewTelegramConfigFromEnv(t *testing.T) {
	t.Run("uses defaults", func(t *testing.T) {
		UnsetEnv(t, telegramEnvVars...)

		cfg := config.NewTelegramConfigFromEnv(testutil.NewDiscardLogger())

		diff := cmp.Diff(defaultTelegramConfig, cfg)
		if diff != "" {
			t.Errorf("NewTelegramConfigFromEnv() diff (-want +got):\n%s", diff)
		}
	})

	t.Run("uses env vars", func(t *testing.T) {
		setCustomTelegramEnvVars(t)

		cfg := config.NewTelegramConfigFromEnv(testutil.NewDiscardLogger())
		wantCfg := config.TelegramConfig{
			Token: secrecy.SecretString("custom_bot_token"),
		}
		diff := cmp.Diff(wantCfg, cfg)
		if diff != "" {
			t.Errorf("NewTelegramConfigFromEnv() diff (-want +got):\n%s", diff)
		}
	})

	t.Run("warnings on unset variables", func(t *testing.T) {
		UnsetEnv(t, telegramEnvVars...)

		logger, buf := testutil.NewBufJsonLogger(t, slog.LevelWarn)
		_ = config.NewTelegramConfigFromEnv(logger)

		for _, envVar := range telegramEnvVars {
			if !strings.Contains(buf.String(), envVar) {
				t.Errorf("got log output %q, want mention of %q", buf.String(), envVar)
			}
		}
	})

	t.Run("no warnings if all variables are set", func(t *testing.T) {
		setCustomTelegramEnvVars(t)

		logger, buf := testutil.NewBufJsonLogger(t, slog.LevelWarn)
		_ = config.NewTelegramConfigFromEnv(logger)

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
		"token": "(10 chars)",
	}

	testutil.FailOnInvalidLogValue(t, v.Group(), wantAttrs)
}
