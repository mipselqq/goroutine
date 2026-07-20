package config

import (
	"log/slog"

	"goroutine/internal/secrecy"
)

type Redis struct {
	Host     string
	Port     string
	Password secrecy.SecretString
}

func NewRedisFromEnv(logger *slog.Logger) Redis {
	return Redis{
		Host:     getEnvStringOrDefault("REDIS_HOST", "127.0.0.1", logger),
		Port:     getEnvStringOrDefault("REDIS_PORT", "6379", logger),
		Password: secrecy.SecretString(getEnvStringOrDefault("REDIS_PASSWORD", "redis_password", logger)),
	}
}

func (c *Redis) BuildAddr() string {
	return c.Host + ":" + c.Port
}

//nolint:gocritic // Pointer receiver disables formatting
func (c Redis) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("host", c.Host),
		slog.String("port", c.Port),
		slog.Any("password", c.Password),
	)
}
