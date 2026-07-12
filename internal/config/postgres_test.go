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

var defaultPgConfig = config.PG{
	User:     "user",
	Password: secrecy.SecretString("password"),
	Host:     "127.0.0.1",
	Port:     "5432",
	DB:       "todo_db",
}

var pgEnvVars = []string{"POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_HOST", "POSTGRES_PORT", "POSTGRES_DB"}

func setCustomPgEnvVars(t *testing.T) {
	t.Setenv("POSTGRES_USER", "custom_user")
	t.Setenv("POSTGRES_PASSWORD", "custom_pass")
	t.Setenv("POSTGRES_HOST", "custom_host")
	t.Setenv("POSTGRES_PORT", "5433")
	t.Setenv("POSTGRES_DB", "custom_db")
}

func TestNewPGFromEnv(t *testing.T) {
	t.Run("uses defaults", func(t *testing.T) {
		UnsetEnv(t, pgEnvVars...)

		cfg := config.NewPGFromEnv(testutil.NewDiscardLogger())

		diff := cmp.Diff(defaultPgConfig, cfg)
		if diff != "" {
			t.Errorf("NewPGFromEnv() diff (-want +got):\n%s", diff)
		}
	})

	t.Run("uses env vars", func(t *testing.T) {
		setCustomPgEnvVars(t)

		cfg := config.NewPGFromEnv(testutil.NewDiscardLogger())
		wantCfg := config.PG{
			User:     "custom_user",
			Password: secrecy.SecretString("custom_pass"),
			Host:     "custom_host",
			Port:     "5433",
			DB:       "custom_db",
		}
		diff := cmp.Diff(wantCfg, cfg)
		if diff != "" {
			t.Errorf("NewPGFromEnv() diff (-want +got):\n%s", diff)
		}
	})

	t.Run("warnings on unset variables", func(t *testing.T) {
		UnsetEnv(t, pgEnvVars...)

		logger, buf := testutil.NewBufJSONLogger(t, slog.LevelWarn)
		_ = config.NewPGFromEnv(logger)

		for _, envVar := range pgEnvVars {
			if !strings.Contains(buf.String(), envVar) {
				t.Errorf("got log output %q, want mention of %q", buf.String(), envVar)
			}
		}
	})

	t.Run("no warnings if all variables are set", func(t *testing.T) {
		setCustomPgEnvVars(t)

		logger, buf := testutil.NewBufJSONLogger(t, slog.LevelWarn)
		_ = config.NewPGFromEnv(logger)

		if buf.String() != "" {
			t.Errorf("got warnings %q, want none", buf.String())
		}
	})
}

func TestPG_BuildDSN(t *testing.T) {
	want := "postgres://user:password@127.0.0.1:5432/todo_db"

	dsn := defaultPgConfig.BuildDSN()
	got := dsn.RevealSecret()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	testutil.AssertSecretHidden(t, want, dsn)
}

func TestPG_LogValue(t *testing.T) {
	v := defaultPgConfig.LogValue()
	if v.Kind() != slog.KindGroup {
		t.Fatalf("got kind %v, want Group", v.Kind())
	}

	wantAttrs := map[string]string{
		"user":     "user",
		"password": "(8 chars)",
		"host":     "127.0.0.1",
		"port":     "5432",
		"db":       "todo_db",
	}

	testutil.FailOnInvalidLogValue(t, v.Group(), wantAttrs)
}
