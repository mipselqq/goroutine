package config

import (
	"log/slog"
	"os"
)

func getenvOrDefault(key, def string, logger *slog.Logger) string {
	env := os.Getenv(key)

	if env == "" {
		logger.Warn("Environment variable not set, using default:", slog.String("key", key))
		return def
	}
	return env
}
