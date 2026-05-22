package config

import (
	"log/slog"

	"goroutine/internal/secrecy"
)

type TelegramConfig struct {
	Token secrecy.SecretString
}

func NewTelegramConfigFromEnv(logger *slog.Logger) TelegramConfig {
	return TelegramConfig{
		Token: secrecy.SecretString(getenvOrDefault("TELEGRAM_BOT_TOKEN", "mock_token", logger)),
	}
}

//nolint:gocritic // Pointer receiver disables formatting
func (c TelegramConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("token", c.Token),
	)
}
