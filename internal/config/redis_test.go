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

var defaultRedisConfig = config.Redis{
	Host:     "127.0.0.1",
	Port:     "6379",
	Password: secrecy.SecretString("redis_password"),
}

var redisEnvVars = []string{"REDIS_HOST", "REDIS_PORT", "REDIS_PASSWORD"}

func setCustomRedisEnvVars(t *testing.T) {
	t.Setenv("REDIS_HOST", "custom_redis_host")
	t.Setenv("REDIS_PORT", "6380")
	t.Setenv("REDIS_PASSWORD", "custom_redis_pass")
}

func TestNewRedisFromEnv(t *testing.T) {
	t.Run("uses defaults", func(t *testing.T) {
		UnsetEnv(t, redisEnvVars...)

		cfg := config.NewRedisFromEnv(testutil.NewDiscardLogger())

		diff := cmp.Diff(defaultRedisConfig, cfg)
		if diff != "" {
			t.Errorf("NewRedisFromEnv() diff (-want +got):\n%s", diff)
		}
	})

	t.Run("uses env vars", func(t *testing.T) {
		setCustomRedisEnvVars(t)

		cfg := config.NewRedisFromEnv(testutil.NewDiscardLogger())
		wantCfg := config.Redis{
			Host:     "custom_redis_host",
			Port:     "6380",
			Password: secrecy.SecretString("custom_redis_pass"),
		}
		diff := cmp.Diff(wantCfg, cfg)
		if diff != "" {
			t.Errorf("NewRedisFromEnv() diff (-want +got):\n%s", diff)
		}
	})

	t.Run("warnings on unset variables", func(t *testing.T) {
		UnsetEnv(t, redisEnvVars...)

		logger, buf := testutil.NewBufJsonLogger(t, slog.LevelWarn)
		_ = config.NewRedisFromEnv(logger)

		for _, envVar := range redisEnvVars {
			if !strings.Contains(buf.String(), envVar) {
				t.Errorf("got log output %q, want mention of %q", buf.String(), envVar)
			}
		}
	})

	t.Run("no warnings if all variables are set", func(t *testing.T) {
		setCustomRedisEnvVars(t)

		logger, buf := testutil.NewBufJsonLogger(t, slog.LevelWarn)
		_ = config.NewRedisFromEnv(logger)

		if buf.String() != "" {
			t.Errorf("got warnings %q, want none", buf.String())
		}
	})
}

func TestRedis_BuildAddr(t *testing.T) {
	want := "127.0.0.1:6379"
	got := defaultRedisConfig.BuildAddr()
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestRedis_LogValue(t *testing.T) {
	v := defaultRedisConfig.LogValue()
	if v.Kind() != slog.KindGroup {
		t.Fatalf("got kind %v, want Group", v.Kind())
	}

	wantAttrs := map[string]string{
		"host":     "127.0.0.1",
		"port":     "6379",
		"password": "(14 chars)",
	}

	testutil.FailOnInvalidLogValue(t, v.Group(), wantAttrs)
}
