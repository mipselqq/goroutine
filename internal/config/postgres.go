package config

import (
	"fmt"
	"log/slog"

	"goroutine/internal/secrecy"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Pg struct {
	User     string
	Password secrecy.SecretString
	Host     string
	Port     string
	DB       string
}

func NewPGFromEnv(logger *slog.Logger) Pg {
	return Pg{
		User:     getEnvStringOrDefault("POSTGRES_USER", "user", logger),
		Password: secrecy.SecretString(getEnvStringOrDefault("POSTGRES_PASSWORD", "password", logger)),
		Host:     getEnvStringOrDefault("POSTGRES_HOST", "127.0.0.1", logger),
		Port:     getEnvStringOrDefault("POSTGRES_PORT", "5432", logger),
		DB:       getEnvStringOrDefault("POSTGRES_DB", "todo_db", logger),
	}
}

func (c *Pg) BuildDSN() secrecy.SecretString {
	return secrecy.SecretString(fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.User, c.Password.RevealSecret(), c.Host, c.Port, c.DB))
}

func (c *Pg) ParsePGXpoolConfig() (*pgxpool.Config, error) {
	config, err := pgxpool.ParseConfig(c.BuildDSN().RevealSecret())
	if err != nil {
		return nil, err
	}

	if config.ConnConfig.RuntimeParams == nil {
		config.ConnConfig.RuntimeParams = map[string]string{}
	}
	config.ConnConfig.RuntimeParams["timezone"] = "UTC"

	return config, nil
}

//nolint:gocritic // Pointer receiver disables formatting
func (c Pg) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("user", c.User),
		slog.Any("password", c.Password),
		slog.String("host", c.Host),
		slog.String("port", c.Port),
		slog.String("db", c.DB),
	)
}
