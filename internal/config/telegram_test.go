package config_test

import (
	"log/slog"
	"strings"
	"testing"
	"time"

	"goroutine/internal/config"
	"goroutine/internal/testutil"

	"github.com/google/go-cmp/cmp"
)

func setCustomTelegramEnvVars(t *testing.T) {
	t.Setenv("TELEGRAM_BOT_TOKEN", testutil.AnotherValidTelegramToken().RevealSecret())
	t.Setenv("TELEGRAM_LINK_TOKEN_TTL", "30m")
}

func TestNewTelegramConfigFromEnv(t *testing.T) {
	t.Run("uses env vars", func(t *testing.T) {
		setCustomTelegramEnvVars(t)

		cfg := MustNewTelegramConfigFromEnv(t)
		wantCfg := config.TelegramConfig{
			Token:        testutil.AnotherValidTelegramToken(),
			LinkTokenTTL: 30 * time.Minute,
		}
		diff := cmp.Diff(wantCfg, cfg, testutil.CmpAllowUnexported())
		if diff != "" {
			t.Errorf("NewTelegramConfigFromEnv() diff (-want +got):\n%s", diff)
		}
	})

	t.Run("invalid bot token format", func(t *testing.T) {
		t.Setenv("TELEGRAM_BOT_TOKEN", "not-a-valid-token")

		_, err := config.NewTelegramConfigFromEnv(testutil.NewDiscardLogger())
		if err == nil {
			t.Fatal("NewTelegramConfigFromEnv() error = nil, want non-nil")
		}
	})

	t.Run("missing bot token", func(t *testing.T) {
		t.Setenv("TELEGRAM_LINK_TOKEN_TTL", "30m")

		_, err := config.NewTelegramConfigFromEnv(testutil.NewDiscardLogger())
		if err == nil {
			t.Fatal("NewTelegramConfigFromEnv() error = nil, want non-nil")
		}
	})

	t.Run("invalid duration", func(t *testing.T) {
		t.Setenv("TELEGRAM_BOT_TOKEN", testutil.ValidTelegramToken().RevealSecret())
		t.Setenv("TELEGRAM_LINK_TOKEN_TTL", "not-a-duration")

		_, err := config.NewTelegramConfigFromEnv(testutil.NewDiscardLogger())
		if err == nil {
			t.Fatal("NewTelegramConfigFromEnv() error = nil, want non-nil")
		}
	})

	t.Run("uses default link token ttl", func(t *testing.T) {
		t.Setenv("TELEGRAM_BOT_TOKEN", testutil.ValidTelegramToken().RevealSecret())

		logger, buf := testutil.NewBufJsonLogger(t, slog.LevelWarn)
		cfg, err := config.NewTelegramConfigFromEnv(logger)
		if err != nil {
			t.Fatalf("NewTelegramConfigFromEnv() error = %v", err)
		}

		if cfg.LinkTokenTTL != 15*time.Minute {
			t.Errorf("LinkTokenTTL = %v, want 15m", cfg.LinkTokenTTL)
		}
		if !strings.Contains(buf.String(), "TELEGRAM_LINK_TOKEN_TTL") {
			t.Errorf("got log output %q, want mention of TELEGRAM_LINK_TOKEN_TTL", buf.String())
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
	cfg := config.TelegramConfig{
		Token:        testutil.ValidTelegramToken(),
		LinkTokenTTL: 15 * time.Minute,
	}

	v := cfg.LogValue()
	if v.Kind() != slog.KindGroup {
		t.Fatalf("got kind %v, want Group", v.Kind())
	}

	wantAttrs := map[string]string{
		"token":          "(46 chars)",
		"link_token_ttl": "15m0s",
	}

	testutil.FailOnInvalidLogValue(t, v.Group(), wantAttrs)
}
