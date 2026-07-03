package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"goroutine/internal/secrecy"
)

type TelegramConfig struct {
	Token        secrecy.SecretString
	LinkTokenTTL time.Duration
}

func NewTelegramConfigFromEnv(logger *slog.Logger) (TelegramConfig, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return TelegramConfig{}, fmt.Errorf("telegram config: TELEGRAM_BOT_TOKEN is required")
	}

	linkTokenTTL, err := getEnvTimeOrDefault("TELEGRAM_LINK_TOKEN_TTL", 15*time.Minute, logger)
	if err != nil {
		return TelegramConfig{}, fmt.Errorf("telegram config: %w", err)
	}

	return TelegramConfig{
		Token:        secrecy.SecretString(token),
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
