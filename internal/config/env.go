package config

import (
	"log/slog"
	"os"
)

func getenvOrDefault(key, def string, logger *slog.Logger) string {
	env := os.Getenv(key)

	if env == "" {
		logger.Warn("Environment variable not set", slog.String("key", key), slog.String("default", def))
		return def
	}
	return env
}
