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
	BaseURL      string
	LinkTokenTTL time.Duration
}

func NewTelegramConfigFromEnv(logger *slog.Logger) (TelegramConfig, error) {
	envToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if envToken == "" {
		return TelegramConfig{}, fmt.Errorf("telegram config: TELEGRAM_BOT_TOKEN is required")
	}

	token, err := domain.NewTelegramToken(envToken)
	if err != nil {
		return TelegramConfig{}, fmt.Errorf("telegram config: %w", err)
	}

	linkTokenTTL, err := getEnvTimeOrDefault("TELEGRAM_LINK_TOKEN_TTL", 15*time.Minute, logger)
	if err != nil {
		return TelegramConfig{}, fmt.Errorf("telegram config: %w", err)
	}

	baseURL := getenvOrDefault("TELEGRAM_API_BASE_URL", "https://api.telegram.org", logger)

	return TelegramConfig{
		Token:        token,
		BaseURL:      baseURL,
		LinkTokenTTL: linkTokenTTL,
	}, nil
}

//nolint:gocritic // Pointer receiver disables formatting
func (c TelegramConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("token", c.Token),
		slog.String("base_url", c.BaseURL),
		slog.Duration("link_token_ttl", c.LinkTokenTTL),
	)
}
