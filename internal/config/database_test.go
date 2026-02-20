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

var defaultPgConfig = config.PgConfig{
	User:     "user",
	Password: secrecy.SecretString("password"),
	Host:     "127.0.0.1",
	Port:     "5432",
	DB:       "todo_db",
}

var pgEnvVars = []string{"POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_DB"}

func TestNewPGConfigFromEnv(t *testing.T) {
	t.Run("uses defaults", func(t *testing.T) {
		testutil.UnsetEnv(t, pgEnvVars...)

		cfg := config.NewPGConfigFromEnv(testutil.NewDiscardLogger())

		diff := cmp.Diff(defaultPgConfig, cfg)
		if diff != "" {
			t.Errorf("invalid defaults (-want +got):\n%s", diff)
		}
	})

	t.Run("uses env vars", func(t *testing.T) {
		t.Setenv("POSTGRES_USER", "custom_user")
		t.Setenv("POSTGRES_PASSWORD", "custom_pass")
		t.Setenv("POSTGRES_HOST", "custom_host")
		t.Setenv("POSTGRES_PORT", "5433")
		t.Setenv("POSTGRES_DB", "custom_db")

		cfg := config.NewPGConfigFromEnv(testutil.NewDiscardLogger())
		expectedCfg := config.PgConfig{
			User:     "custom_user",
			Password: secrecy.SecretString("custom_pass"),
			Host:     "custom_host",
			Port:     "5433",
			DB:       "custom_db",
		}
		diff := cmp.Diff(expectedCfg, cfg)
		if diff != "" {
			t.Errorf("invalid defaults (-want +got):\n%s", diff)
		}
	})

	t.Run("warns unset variables", func(t *testing.T) {
		logger, buf := testutil.NewBufJsonLogger(t)
		_ = config.NewPGConfigFromEnv(logger)

		for _, envVar := range pgEnvVars {
			if !strings.Contains(buf.String(), envVar) {
				t.Errorf("expected warn on unset %s", envVar)
			}
		}
	})
}

func TestPgConfig_BuildDSN(t *testing.T) {
	expected := "postgres://user:password@127.0.0.1:5432/todo_db"
	if got := defaultPgConfig.BuildDSN(); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestPgConfig_LogValue(t *testing.T) {
	v := defaultPgConfig.LogValue()
	if v.Kind() != slog.KindGroup {
		t.Fatalf("expected Group kind, got %v", v.Kind())
	}

	expectedAttrs := map[string]string{
		"user":     "user",
		"password": "(8 chars)",
		"host":     "127.0.0.1",
		"port":     "5432",
		"db":       "todo_db",
	}

	testutil.FailOnInvalidLogValue(t, v.Group(), expectedAttrs)
}
