package config

import (
	"fmt"
	"log/slog"

	"goroutine/internal/secrecy"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgConfig struct {
	user     string
	password secrecy.SecretString
	host     string
	port     string
	db       string
}

func NewPGConfigFromEnv() PgConfig {
	return PgConfig{
		user:     getenvOrDefault("POSTGRES_USER", "user"),
		password: secrecy.SecretString(getenvOrDefault("POSTGRES_PASSWORD", "password")),
		host:     getenvOrDefault("POSTGRES_HOST", "127.0.0.1"),
		port:     getenvOrDefault("POSTGRES_PORT", "5432"),
		db:       getenvOrDefault("POSTGRES_DB", "todo_db"),
	}
}

func (c *PgConfig) buildDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.user, c.password.RevealSecret(), c.host, c.port, c.db)
}

func (c *PgConfig) ParsePGXpoolConfig() (*pgxpool.Config, error) {
	config, err := pgxpool.ParseConfig(c.buildDSN())
	if err != nil {
		return nil, err
	}

	return config, nil
}

//nolint:gocritic // Pointer receiver disables formatting
func (c PgConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("user", c.user),
		slog.Any("password", c.password),
		slog.String("host", c.host),
		slog.String("port", c.port),
		slog.String("db", c.db),
	)
}
