package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"
)

func getEnvStringOrDefault(key, def string, logger *slog.Logger) string {
	env := os.Getenv(key)

	if env == "" {
		logger.Warn("Environment variable not set, using default:", slog.String("key", key))
		return def
	}
	return env
}

func getEnvDurationOrDefault(key string, def time.Duration, logger *slog.Logger) (time.Duration, error) {
	env := os.Getenv(key)

	if env == "" {
		logger.Warn("Environment variable not set, using default:", slog.String("key", key))
		return def, nil
	}

	val, err := time.ParseDuration(env)
	if err != nil {
		return 0, fmt.Errorf("invalid duration for %s=%q: %w", key, env, err)
	}

	return val, nil
}
