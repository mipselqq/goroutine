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
	t.Setenv("TELEGRAM_API_BASE_URL", "https://custom.telegram.org")
}

func TestNewTelegramFromEnv(t *testing.T) {
	t.Run("uses env vars", func(t *testing.T) {
		setCustomTelegramEnvVars(t)

		cfg := MustNewTelegramFromEnv(t)
		wantCfg := config.Telegram{
			Token:        testutil.AnotherValidTelegramToken(),
			BaseURL:      "https://custom.telegram.org",
			LinkTokenTTL: 30 * time.Minute,
		}
		diff := cmp.Diff(wantCfg, cfg, testutil.CmpAllowUnexported())
		if diff != "" {
			t.Errorf("NewTelegramFromEnv() diff (-want +got):\n%s", diff)
		}
	})

	t.Run("invalid bot token format", func(t *testing.T) {
		t.Setenv("TELEGRAM_BOT_TOKEN", "not-a-valid-token")

		_, err := config.NewTelegramFromEnv(testutil.NewDiscardLogger())
		if err == nil {
			t.Fatal("NewTelegramFromEnv() error = nil, want non-nil")
		}
	})

	t.Run("missing bot token", func(t *testing.T) {
		t.Setenv("TELEGRAM_LINK_TOKEN_TTL", "30m")

		_, err := config.NewTelegramFromEnv(testutil.NewDiscardLogger())
		if err == nil {
			t.Fatal("NewTelegramFromEnv() error = nil, want non-nil")
		}
	})

	t.Run("invalid duration", func(t *testing.T) {
		t.Setenv("TELEGRAM_BOT_TOKEN", testutil.ValidTelegramToken().RevealSecret())
		t.Setenv("TELEGRAM_LINK_TOKEN_TTL", "not-a-duration")

		_, err := config.NewTelegramFromEnv(testutil.NewDiscardLogger())
		if err == nil {
			t.Fatal("NewTelegramFromEnv() error = nil, want non-nil")
		}
	})

	t.Run("uses default link token ttl", func(t *testing.T) {
		t.Setenv("TELEGRAM_BOT_TOKEN", testutil.ValidTelegramToken().RevealSecret())

		logger, buf := testutil.NewBufJsonLogger(t, slog.LevelWarn)
		cfg, err := config.NewTelegramFromEnv(logger)
		if err != nil {
			t.Fatalf("NewTelegramFromEnv() error = %v", err)
		}

		if cfg.LinkTokenTTL != 15*time.Minute {
			t.Errorf("LinkTokenTTL = %v, want 15m", cfg.LinkTokenTTL)
		}
		if !strings.Contains(buf.String(), "TELEGRAM_LINK_TOKEN_TTL") {
			t.Errorf("got log output %q, want mention of TELEGRAM_LINK_TOKEN_TTL", buf.String())
		}
	})

	t.Run("uses custom base url", func(t *testing.T) {
		t.Setenv("TELEGRAM_BOT_TOKEN", testutil.ValidTelegramToken().RevealSecret())
		t.Setenv("TELEGRAM_API_BASE_URL", "http://localhost:9999")

		cfg := MustNewTelegramFromEnv(t)
		if cfg.BaseURL != "http://localhost:9999" {
			t.Errorf("got BaseURL %q, want http://localhost:9999", cfg.BaseURL)
		}
	})

	t.Run("uses default base url", func(t *testing.T) {
		t.Setenv("TELEGRAM_BOT_TOKEN", testutil.ValidTelegramToken().RevealSecret())

		cfg := MustNewTelegramFromEnv(t)
		if cfg.BaseURL != "https://api.telegram.org" {
			t.Errorf("got BaseURL %q, want https://api.telegram.org", cfg.BaseURL)
		}
	})

	t.Run("no warnings if all variables are set", func(t *testing.T) {
		setCustomTelegramEnvVars(t)

		logger, buf := testutil.NewBufJsonLogger(t, slog.LevelWarn)
		_, err := config.NewTelegramFromEnv(logger)
		if err != nil {
			t.Fatalf("NewTelegramFromEnv() error = %v", err)
		}

		if buf.String() != "" {
			t.Errorf("got warnings %q, want none", buf.String())
		}
	})
}

func TestTelegram_LogValue(t *testing.T) {
	cfg := config.Telegram{
		Token:        testutil.ValidTelegramToken(),
		BaseURL:      "https://api.telegram.org",
		LinkTokenTTL: 15 * time.Minute,
	}

	v := cfg.LogValue()
	if v.Kind() != slog.KindGroup {
		t.Fatalf("got kind %v, want Group", v.Kind())
	}

	wantAttrs := map[string]string{
		"token":          "(46 chars)",
		"base_url":       "https://api.telegram.org",
		"link_token_ttl": "15m0s",
	}

	testutil.FailOnInvalidLogValue(t, v.Group(), wantAttrs)
}
