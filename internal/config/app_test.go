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

var defaultAppConfig = config.App{
	Port:           "8080",
	AdminPort:      "9091",
	Host:           "0.0.0.0",
	SwaggerHost:    "localhost:8080",
	LogLevel:       "info",
	Env:            "dev",
	JWTSecret:      secrecy.SecretString("very_secret"),
	JWTExp:         24 * time.Hour,
	AllowedOrigins: config.ParseAllowedOrigins("http://localhost:8080,http://127.0.0.1:8080"),
}

var appEnvVars = []string{"PORT", "ADMIN_PORT", "HOST", "SWAGGER_HOST", "LOG_LEVEL", "ENV", "JWT_SECRET", "JWT_EXP", "ALLOWED_ORIGINS"}

func setCustomAppEnvVars(t *testing.T) {
	t.Setenv("PORT", "3000")
	t.Setenv("HOST", "127.0.0.1")
	t.Setenv("ADMIN_PORT", "9092")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("ENV", "prod")
	t.Setenv("SWAGGER_HOST", "example.com")
	t.Setenv("JWT_SECRET", "more_secret")
	t.Setenv("JWT_EXP", "1h")
	t.Setenv("ALLOWED_ORIGINS", "http://example.com,http://test.com")
}

func TestNewAppFromEnv(t *testing.T) {
	t.Run("uses defaults", func(t *testing.T) {
		UnsetEnv(t, appEnvVars...)

		cfg := config.NewAppFromEnv(testutil.NewDiscardLogger())

		diff := cmp.Diff(defaultAppConfig, cfg)
		if diff != "" {
			t.Errorf("NewAppFromEnv() diff (-want +got):\n%s", diff)
		}
	})

	t.Run("uses env vars", func(t *testing.T) {
		setCustomAppEnvVars(t)

		cfg := config.NewAppFromEnv(testutil.NewDiscardLogger())
		wantCfg := config.App{
			Port:           "3000",
			AdminPort:      "9092",
			Host:           "127.0.0.1",
			SwaggerHost:    "example.com",
			LogLevel:       "debug",
			Env:            "prod",
			JWTSecret:      secrecy.SecretString("more_secret"),
			JWTExp:         time.Hour,
			AllowedOrigins: config.ParseAllowedOrigins("http://example.com,http://test.com"),
		}
		diff := cmp.Diff(wantCfg, cfg)
		if diff != "" {
			t.Errorf("NewAppFromEnv() diff (-want +got):\n%s", diff)
		}
	})

	t.Run("warnings on unset variables", func(t *testing.T) {
		UnsetEnv(t, appEnvVars...)
		logger, buf := testutil.NewBufJSONLogger(t, slog.LevelWarn)
		_ = config.NewAppFromEnv(logger)

		for _, envVar := range appEnvVars {
			if !strings.Contains(buf.String(), envVar) {
				t.Errorf("got log output %q, want mention of %q", buf.String(), envVar)
			}
		}
	})

	t.Run("no warnings on all variables are set", func(t *testing.T) {
		setCustomAppEnvVars(t)

		logger, buf := testutil.NewBufJSONLogger(t, slog.LevelWarn)
		_ = config.NewAppFromEnv(logger)

		if buf.String() != "" {
			t.Errorf("got warnings %q, want none", buf.String())
		}
	})
}

func TestApp_LogValue(t *testing.T) {
	v := defaultAppConfig.LogValue()
	if v.Kind() != slog.KindGroup {
		t.Fatalf("got kind %v, want Group", v.Kind())
	}

	attrs := v.Group()
	wantAttrs := map[string]string{
		"port":            "8080",
		"admin_port":      "9091",
		"host":            "0.0.0.0",
		"log_level":       "info",
		"env":             "dev",
		"swagger_host":    "localhost:8080",
		"jwt_secret":      "(11 chars)",
		"jwt_exp":         "24h0m0s",
		"allowed_origins": "[http://127.0.0.1:8080 http://localhost:8080]",
	}

	testutil.FailOnInvalidLogValue(t, attrs, wantAttrs)
}
