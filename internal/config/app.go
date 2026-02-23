package config

import (
	"log/slog"
	"sort"
	"strings"
	"time"

	"goroutine/internal/secrecy"
)

type AppConfig struct {
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

func NewAppConfigFromEnv(logger *slog.Logger) AppConfig {
	jwtExpStr := getenvOrDefault("JWT_EXP", "24h", logger)
	jwtExp, err := time.ParseDuration(jwtExpStr)
	if err != nil {
		jwtExp = 24 * time.Hour
	}

	allowedOrigins := getenvOrDefault("ALLOWED_ORIGINS", "http://localhost:8080,http://127.0.0.1:8080", logger)
	return AppConfig{
		Port:           getenvOrDefault("PORT", "8080", logger),
		AdminPort:      getenvOrDefault("ADMIN_PORT", "9091", logger),
		Host:           getenvOrDefault("HOST", "0.0.0.0", logger),
		LogLevel:       getenvOrDefault("LOG_LEVEL", "info", logger),
		Env:            getenvOrDefault("ENV", "dev", logger),
		SwaggerHost:    getenvOrDefault("SWAGGER_HOST", "localhost:8080", logger),
		JWTSecret:      secrecy.SecretString(getenvOrDefault("JWT_SECRET", "very_secret", logger)),
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
func (c AppConfig) LogValue() slog.Value {
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
		slog.String("jwt_secret", c.JWTSecret.String()),
		slog.Duration("jwt_exp", c.JWTExp),
		slog.Any("allowed_origins", allowedOrigins),
	)
}
