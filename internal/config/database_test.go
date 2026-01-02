package config

import (
	"log/slog"
	"testing"
)

func TestNewPGConfigFromEnv(t *testing.T) {
	t.Run("uses defaults", func(t *testing.T) {
		t.Setenv("POSTGRES_USER", "")
		t.Setenv("POSTGRES_PASSWORD", "")
		t.Setenv("POSTGRES_HOST", "")
		t.Setenv("POSTGRES_PORT", "")
		t.Setenv("POSTGRES_DB", "")

		cfg := NewPGConfigFromEnv()
		if cfg.user != "user" {
			t.Errorf("expected default user 'user', got %q", cfg.user)
		}
		if cfg.port != "5432" {
			t.Errorf("expected default port '5432', got %q", cfg.port)
		}
	})

	t.Run("uses env vars", func(t *testing.T) {
		t.Setenv("POSTGRES_USER", "custom_user")
		t.Setenv("POSTGRES_PASSWORD", "custom_pass")
		t.Setenv("POSTGRES_HOST", "custom_host")
		t.Setenv("POSTGRES_PORT", "5433")
		t.Setenv("POSTGRES_DB", "custom_db")

		cfg := NewPGConfigFromEnv()
		if cfg.user != "custom_user" {
			t.Errorf("got %s", cfg.user)
		}
		if cfg.password != "custom_pass" {
			t.Errorf("got %s", cfg.password)
		}
		if cfg.host != "custom_host" {
			t.Errorf("got %s", cfg.host)
		}
		if cfg.port != "5433" {
			t.Errorf("got %s", cfg.port)
		}
		if cfg.db != "custom_db" {
			t.Errorf("got %s", cfg.db)
		}
	})
}

func TestPgConfig_BuildDSN(t *testing.T) {
	cfg := PgConfig{
		user:     "user",
		password: "password",
		host:     "localhost",
		port:     "5432",
		db:       "dbname",
	}

	expected := "postgres://user:password@localhost:5432/dbname"
	if got := cfg.buildDSN(); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestPgConfig_LogValue(t *testing.T) {
	cfg := PgConfig{
		user:     "u",
		password: "super_secret_password",
		host:     "h",
		port:     "p",
		db:       "d",
	}

	v := cfg.LogValue()
	if v.Kind() != slog.KindGroup {
		t.Fatalf("expected Group kind, got %v", v.Kind())
	}

	attrs := v.Group()
	foundPassword := false
	for _, a := range attrs {
		if a.Key == "password" {
			foundPassword = true
			if a.Value.String() == "super_secret_password" {
				t.Error("password was not masked")
			}
			expected := "(21 chars)"
			if a.Value.String() != expected {
				t.Errorf("expected masked password %q, got %q", expected, a.Value.String())
			}
		}
	}
	if !foundPassword {
		t.Error("password field not found in log value")
	}
}
