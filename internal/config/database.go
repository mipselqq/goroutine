package config

import (
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgConfig struct {
	user     string
	password string
	host     string
	port     string
	db       string
}

func NewPGConfigFromEnv() PgConfig {
	return PgConfig{
		user:     getenvOrDefault("POSTGRES_USER", "user"),
		password: getenvOrDefault("POSTGRES_PASSWORD", "password"),
		host:     getenvOrDefault("POSTGRES_HOST", "127.0.0.1"),
		port:     getenvOrDefault("POSTGRES_PORT", "5432"),
		db:       getenvOrDefault("POSTGRES_DB", "todo_db"),
	}
}

func (c *PgConfig) buildDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.user, c.password, c.host, c.port, c.db)
}

func (c *PgConfig) ParsePGXpoolConfig() (*pgxpool.Config, error) {
	config, err := pgxpool.ParseConfig(c.buildDSN())
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *PgConfig) LogValue() slog.Value {
	hiddenPassword := fmt.Sprintf("(%d chars)", len(c.password))

	return slog.GroupValue(
		slog.String("user", c.user),
		slog.String("password", hiddenPassword),
		slog.String("host", c.host),
		slog.String("port", c.port),
		slog.String("db", c.db),
	)
}
