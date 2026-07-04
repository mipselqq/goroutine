package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"goroutine/internal/domain"
)

type TelegramConfig struct {
	Token        domain.TelegramToken
	LinkTokenTTL time.Duration
}

func NewTelegramConfigFromEnv(logger *slog.Logger) (TelegramConfig, error) {
	env_token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if env_token == "" {
		return TelegramConfig{}, fmt.Errorf("telegram config: TELEGRAM_BOT_TOKEN is required")
	}

	token, err := domain.NewTelegramToken(env_token)
	if err != nil {
		return TelegramConfig{}, fmt.Errorf("telegram config: %w", err)
	}

	linkTokenTTL, err := getEnvTimeOrDefault("TELEGRAM_LINK_TOKEN_TTL", 15*time.Minute, logger)
	if err != nil {
		return TelegramConfig{}, fmt.Errorf("telegram config: %w", err)
	}

	return TelegramConfig{
		Token:        token,
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
