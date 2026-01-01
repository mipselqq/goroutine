package config

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PgConfig struct {
	user     string
	password string
	host     string
	port     string
	db       string
}

func NewPgConfigFromEnv() PgConfig {
	return PgConfig{
		user:     getenvOrDefault("POSTGRES_USER", "user"),
		password: getenvOrDefault("POSTGRES_PASSWORD", "password"),
		host:     getenvOrDefault("POSTGRES_HOST", "localhost"),
		port:     getenvOrDefault("POSTGRES_PORT", "5432"),
		db:       getenvOrDefault("POSTGRES_DB", "todo_db"),
	}
}

func (c PgConfig) buildDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", c.user, c.password, c.host, c.port, c.db)
}

func (c PgConfig) ParsePgxpoolConfig() (*pgxpool.Config, error) {
	config, err := pgxpool.ParseConfig(c.buildDSN())
	if err != nil {
		return nil, err
	}

	return config, nil
}
