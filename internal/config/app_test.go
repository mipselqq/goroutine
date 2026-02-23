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

var defaultAppConfig = config.AppConfig{
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

func TestNewAppConfigFromEnv(t *testing.T) {
	t.Run("uses defaults", func(t *testing.T) {
		testutil.UnsetEnv(t, appEnvVars...)

		cfg := config.NewAppConfigFromEnv(testutil.NewDiscardLogger())

		diff := cmp.Diff(defaultAppConfig, cfg)
		if diff != "" {
			t.Errorf("app used unexpected values instead of defaults (-want +got):\n%s", diff)
		}
	})

	t.Run("uses env vars", func(t *testing.T) {
		setCustomAppEnvVars(t)

		cfg := config.NewAppConfigFromEnv(testutil.NewDiscardLogger())
		expectedCfg := config.AppConfig{
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
		diff := cmp.Diff(expectedCfg, cfg)
		if diff != "" {
			t.Errorf("app config used unexpected values instead of env vars (-want +got):\n%s", diff)
		}
	})

	t.Run("warnings on unset variables", func(t *testing.T) {
		testutil.UnsetEnv(t, appEnvVars...)
		logger, buf := testutil.NewBufJsonLogger(t)
		_ = config.NewAppConfigFromEnv(logger)

		for _, envVar := range appEnvVars {
			if !strings.Contains(buf.String(), envVar) {
				t.Errorf("expected warn on unset %s", envVar)
			}
		}
	})

	t.Run("no warnings on all variables are set", func(t *testing.T) {
		setCustomAppEnvVars(t)

		logger, buf := testutil.NewBufJsonLogger(t)
		_ = config.NewAppConfigFromEnv(logger)

		if buf.String() != "" {
			t.Errorf("expected no warnings on all variables are set, got %s", buf.String())
		}
	})
}

func TestAppConfig_LogValue(t *testing.T) {
	v := defaultAppConfig.LogValue()
	if v.Kind() != slog.KindGroup {
		t.Fatalf("expected Group kind, got %v", v.Kind())
	}

	attrs := v.Group()
	expectedAttrs := map[string]string{
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

	testutil.FailOnInvalidLogValue(t, attrs, expectedAttrs)
}
