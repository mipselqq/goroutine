package config

import (
	"log/slog"

	"goroutine/internal/secrecy"
)

type RedisConfig struct {
	Host     string
	Port     string
	Password secrecy.SecretString
}

func NewRedisFromEnv(logger *slog.Logger) RedisConfig {
	return RedisConfig{
		Host:     getenvOrDefault("REDIS_HOST", "127.0.0.1", logger),
		Port:     getenvOrDefault("REDIS_PORT", "6379", logger),
		Password: secrecy.SecretString(getenvOrDefault("REDIS_PASSWORD", "redis_password", logger)),
	}
}

func (c *RedisConfig) BuildAddr() string {
	return c.Host + ":" + c.Port
}

//nolint:gocritic // Pointer receiver disables formatting
func (c RedisConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("host", c.Host),
		slog.String("port", c.Port),
		slog.Any("password", c.Password),
	)
}
