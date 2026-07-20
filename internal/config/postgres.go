package config

import (
	"fmt"
	"log/slog"

	"goroutine/internal/logging"
	"goroutine/internal/secrecy"
)

type PG struct {
	User     string
	Password secrecy.SecretString
	Host     string
	Port     string
	DB       string
}

func NewPGFromEnv(logger *slog.Logger) PG {
	logger = logging.WithModule(logger, "config.postgres")

	return PG{
		User:     getEnvStringOrDefault("POSTGRES_USER", "user", logger),
		Password: secrecy.SecretString(getEnvStringOrDefault("POSTGRES_PASSWORD", "password", logger)),
		Host:     getEnvStringOrDefault("POSTGRES_HOST", "127.0.0.1", logger),
		Port:     getEnvStringOrDefault("POSTGRES_PORT", "5432", logger),
		DB:       getEnvStringOrDefault("POSTGRES_DB", "todo_db", logger),
	}
}

func (c *PG) BuildDSN() secrecy.SecretString {
	return secrecy.SecretString(fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.User, c.Password.RevealSecret(), c.Host, c.Port, c.DB))
}

//nolint:gocritic // Pointer receiver disables formatting
func (c PG) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("user", c.User),
		slog.Any("password", c.Password),
		slog.String("host", c.Host),
		slog.String("port", c.Port),
		slog.String("db", c.DB),
	)
}
