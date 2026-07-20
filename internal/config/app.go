// Package config loads application and database settings from environment variables.
package config

import (
	"log/slog"
	"sort"
	"strings"
	"time"

	"goroutine/internal/logging"
	"goroutine/internal/secrecy"
)

type App struct {
	Port           string
	AdminPort      string
	Host           string
	LogLevel       string
	Env            string
	SwaggerHost    string
	JWTSecret      secrecy.SecretString
	JWTExp         time.Duration
	AllowedOrigins map[string]struct{}
}

func NewAppFromEnv(logger *slog.Logger) App {
	logger = logging.WithModule(logger, "config.app")

	jwtExpStr := getEnvStringOrDefault("JWT_EXP", "24h", logger)
	jwtExp, err := time.ParseDuration(jwtExpStr)
	if err != nil {
		jwtExp = 24 * time.Hour
	}

	allowedOrigins := getEnvStringOrDefault("ALLOWED_ORIGINS", "http://localhost:8080,http://127.0.0.1:8080", logger)
	return App{
		Port:           getEnvStringOrDefault("PORT", "8080", logger),
		AdminPort:      getEnvStringOrDefault("ADMIN_PORT", "9091", logger),
		Host:           getEnvStringOrDefault("HOST", "0.0.0.0", logger),
		LogLevel:       getEnvStringOrDefault("LOG_LEVEL", "info", logger),
		Env:            getEnvStringOrDefault("ENV", "dev", logger),
		SwaggerHost:    getEnvStringOrDefault("SWAGGER_HOST", "localhost:8080", logger),
		JWTSecret:      secrecy.SecretString(getEnvStringOrDefault("JWT_SECRET", "very_secret", logger)),
		JWTExp:         jwtExp,
		AllowedOrigins: ParseAllowedOrigins(allowedOrigins),
	}
}

func ParseAllowedOrigins(origins string) map[string]struct{} {
	allowedOrigins := strings.Split(origins, ",")
	allowedOriginsMap := make(map[string]struct{})

	for _, origin := range allowedOrigins {
		trimmed := strings.TrimSpace(origin)

		endsOrStartsWithComma := trimmed == ""
		if endsOrStartsWithComma {
			continue
		}

		allowedOriginsMap[trimmed] = struct{}{}
	}

	return allowedOriginsMap
}

//nolint:gocritic // Pointer receiver disables formatting
func (c App) LogValue() slog.Value {
	allowedOrigins := make([]string, 0, len(c.AllowedOrigins))
	for origin := range c.AllowedOrigins {
		allowedOrigins = append(allowedOrigins, origin)
	}
	sort.Strings(allowedOrigins)

	return slog.GroupValue(
		slog.String("port", c.Port),
		slog.String("admin_port", c.AdminPort),
		slog.String("host", c.Host),
		slog.String("log_level", c.LogLevel),
		slog.String("env", c.Env),
		slog.String("swagger_host", c.SwaggerHost),
		slog.Any("jwt_secret", c.JWTSecret),
		slog.Duration("jwt_exp", c.JWTExp),
		slog.Any("allowed_origins", allowedOrigins),
	)
}
