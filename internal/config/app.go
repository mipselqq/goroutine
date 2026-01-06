package config

import (
	"log/slog"

	"go-todo/internal/secrecy"
)

type AppConfig struct {
	Port      string
	LogLevel  string
	Env       string
	JWTSecret secrecy.SecretString
}

func NewAppConfigFromEnv() AppConfig {
	return AppConfig{
		Port:      getenvOrDefault("PORT", "8080"),
		LogLevel:  getenvOrDefault("LOG_LEVEL", "info"),
		Env:       getenvOrDefault("ENV", "dev"),
		JWTSecret: secrecy.SecretString(getenvOrDefault("JWT_SECRET", "very_secret")),
	}
}

//nolint:gocritic // Pointer receiver disables formatting
func (c AppConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("port", c.Port),
		slog.String("log_level", c.LogLevel),
		slog.String("env", c.Env),
		slog.Any("jwt_secret", c.JWTSecret),
	)
}
