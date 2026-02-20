package config

import (
	"fmt"
	"log/slog"

	"goroutine/internal/secrecy"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgConfig struct {
	User     string
	Password secrecy.SecretString
	Host     string
	Port     string
	DB       string
}

func NewPGConfigFromEnv(logger *slog.Logger) PgConfig {
	return PgConfig{
		User:     getenvOrDefault("POSTGRES_USER", "user", logger),
		Password: secrecy.SecretString(getenvOrDefault("POSTGRES_PASSWORD", "password", logger)),
		Host:     getenvOrDefault("POSTGRES_HOST", "127.0.0.1", logger),
		Port:     getenvOrDefault("POSTGRES_PORT", "5432", logger),
		DB:       getenvOrDefault("POSTGRES_DB", "todo_db", logger),
	}
}

func (c *PgConfig) BuildDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.User, c.Password.RevealSecret(), c.Host, c.Port, c.DB)
}

func (c *PgConfig) ParsePGXpoolConfig() (*pgxpool.Config, error) {
	config, err := pgxpool.ParseConfig(c.BuildDSN())
	if err != nil {
		return nil, err
	}

	return config, nil
}

//nolint:gocritic // Pointer receiver disables formatting
func (c PgConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("user", c.User),
		slog.Any("password", c.Password),
		slog.String("host", c.Host),
		slog.String("port", c.Port),
		slog.String("db", c.DB),
	)
}
