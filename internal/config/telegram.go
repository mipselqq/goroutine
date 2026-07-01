package config

import (
	"fmt"
	"log/slog"
	"time"

	"goroutine/internal/secrecy"
)

type TelegramConfig struct {
	Token        secrecy.SecretString
	LinkTokenTTL time.Duration
}

func NewTelegramConfigFromEnv(logger *slog.Logger) (TelegramConfig, error) {
	linkTokenTTL, err := getEnvTimeOrDefault("TELEGRAM_LINK_TOKEN_TTL", 15*time.Minute, logger)
	if err != nil {
		return TelegramConfig{}, fmt.Errorf("telegram config: %w", err)
	}

	return TelegramConfig{
		Token:        secrecy.SecretString(getenvOrDefault("TELEGRAM_BOT_TOKEN", "mock_token", logger)),
		LinkTokenTTL: linkTokenTTL,
	}, nil
}

//nolint:gocritic // Pointer receiver disables formatting
func (c TelegramConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("token", c.Token),
		slog.Duration("link_token_ttl", c.LinkTokenTTL),
	)
}
