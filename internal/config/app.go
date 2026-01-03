package config

import "log/slog"

type AppConfig struct {
	Port     string
	LogLevel string
	Env      string
}

func NewAppConfigFromEnv() AppConfig {
	return AppConfig{
		Port:     getenvOrDefault("PORT", "8080"),
		LogLevel: getenvOrDefault("LOG_LEVEL", "info"),
		Env:      getenvOrDefault("ENV", "dev"),
	}
}

func (c AppConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("port", c.Port),
		slog.String("log_level", c.LogLevel),
		slog.String("env", c.Env),
	)
}
