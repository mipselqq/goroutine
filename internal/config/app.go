package config

import (
	"log/slog"
	"time"

	"goroutine/internal/secrecy"
)

type AppConfig struct {
	Port        string
	AdminPort   string
	Host        string
	LogLevel    string
	Env         string
	SwaggerHost string
	JWTSecret   secrecy.SecretString
	JWTExp      time.Duration
}

func NewAppConfigFromEnv() AppConfig {
	jwtExpStr := getenvOrDefault("JWT_EXP", "24h")
	jwtExp, err := time.ParseDuration(jwtExpStr)
	if err != nil {
		jwtExp = 24 * time.Hour
	}

	return AppConfig{
		Port:        getenvOrDefault("PORT", "8080"),
		AdminPort:   getenvOrDefault("ADMIN_PORT", "9090"),
		Host:        getenvOrDefault("HOST", "0.0.0.0"),
		LogLevel:    getenvOrDefault("LOG_LEVEL", "info"),
		Env:         getenvOrDefault("ENV", "dev"),
		SwaggerHost: getenvOrDefault("SWAGGER_HOST", "localhost:8080"),
		JWTSecret:   secrecy.SecretString(getenvOrDefault("JWT_SECRET", "very_secret")),
		JWTExp:      jwtExp,
	}
}

//nolint:gocritic // Pointer receiver disables formatting
func (c AppConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("port", c.Port),
		slog.String("host", c.Host),
		slog.String("log_level", c.LogLevel),
		slog.String("env", c.Env),
		slog.String("swagger_host", c.SwaggerHost),
		slog.String("jwt_secret", c.JWTSecret.String()),
		slog.Duration("jwt_exp", c.JWTExp),
	)
}
