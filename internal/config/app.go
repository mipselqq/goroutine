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

func NewAppConfigFromEnv(logger *slog.Logger) AppConfig {
	jwtExpStr := getenvOrDefault("JWT_EXP", "24h", logger)
	jwtExp, err := time.ParseDuration(jwtExpStr)
	if err != nil {
		jwtExp = 24 * time.Hour
	}

	return AppConfig{
		Port:        getenvOrDefault("PORT", "8080", logger),
		AdminPort:   getenvOrDefault("ADMIN_PORT", "9091", logger),
		Host:        getenvOrDefault("HOST", "0.0.0.0", logger),
		LogLevel:    getenvOrDefault("LOG_LEVEL", "info", logger),
		Env:         getenvOrDefault("ENV", "dev", logger),
		SwaggerHost: getenvOrDefault("SWAGGER_HOST", "localhost:8080", logger),
		JWTSecret:   secrecy.SecretString(getenvOrDefault("JWT_SECRET", "very_secret", logger)),
		JWTExp:      jwtExp,
	}
}

//nolint:gocritic // Pointer receiver disables formatting
func (c AppConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("port", c.Port),
		slog.String("admin_port", c.AdminPort),
		slog.String("host", c.Host),
		slog.String("log_level", c.LogLevel),
		slog.String("env", c.Env),
		slog.String("swagger_host", c.SwaggerHost),
		slog.String("jwt_secret", c.JWTSecret.String()),
		slog.Duration("jwt_exp", c.JWTExp),
	)
}
