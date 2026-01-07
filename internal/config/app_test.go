package config

import (
	"log/slog"
	"testing"
)

func TestNewAppConfigFromEnv(t *testing.T) {
	t.Run("uses defaults", func(t *testing.T) {
		t.Setenv("PORT", "")
		t.Setenv("LOG_LEVEL", "")
		t.Setenv("ENV", "")
		t.Setenv("JWT_SECRET", "")
		t.Setenv("HOST", "")

		cfg := NewAppConfigFromEnv()
		if cfg.Port != "8080" {
			t.Errorf("expected default port '8080', got %q", cfg.Port)
		}
		if cfg.LogLevel != "info" {
			t.Errorf("expected default log_level 'info', got %q", cfg.LogLevel)
		}
		if cfg.Env != "dev" {
			t.Errorf("expected default env 'dev', got %q", cfg.Env)
		}
		if cfg.JWTSecret.RevealSecret() != "very_secret" {
			t.Errorf("expected default jwt_secret 'very_secret', got %q", string(cfg.JWTSecret))
		}
		if cfg.Host != "0.0.0.0" {
			t.Errorf("expected default host '0.0.0.0', got %q", cfg.Host)
		}
	})

	t.Run("uses env vars", func(t *testing.T) {
		t.Setenv("PORT", "3000")
		t.Setenv("LOG_LEVEL", "debug")
		t.Setenv("ENV", "prod")

		cfg := NewAppConfigFromEnv()
		if cfg.Port != "3000" {
			t.Errorf("expected port '3000', got %q", cfg.Port)
		}
		if cfg.LogLevel != "debug" {
			t.Errorf("expected log_level 'debug', got %q", cfg.LogLevel)
		}
		if cfg.Env != "prod" {
			t.Errorf("expected env 'prod', got %q", cfg.Env)
		}
	})
}

func TestAppConfig_LogValue(t *testing.T) {
	cfg := AppConfig{
		Port:      "8080",
		Host:      "localhost",
		LogLevel:  "info",
		Env:       "dev",
		JWTSecret: "secret",
	}

	v := cfg.LogValue()
	if v.Kind() != slog.KindGroup {
		t.Fatalf("expected Group kind, got %v", v.Kind())
	}

	attrs := v.Group()
	expectedAttrs := map[string]string{
		"port":       "8080",
		"host":       "localhost",
		"log_level":  "info",
		"env":        "dev",
		"jwt_secret": "(6 chars)",
	}

	for _, a := range attrs {
		expected, ok := expectedAttrs[a.Key]
		if !ok {
			t.Errorf("unexpected attribute %q", a.Key)
			continue
		}
		if a.Value.String() != expected {
			t.Errorf("for key %q, expected %q, got %q", a.Key, expected, a.Value.String())
		}
		delete(expectedAttrs, a.Key)
	}

	for key := range expectedAttrs {
		t.Errorf("missing attribute %q", key)
	}
}
