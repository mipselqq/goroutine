package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/logging"
)

type Telegram struct {
	Token        domain.TelegramToken
	BaseURL      string
	LinkTokenTTL time.Duration
}

func NewTelegramFromEnv(logger *slog.Logger) (Telegram, error) {
	logger = logging.WithModule(logger, "config.telegram")

	envToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if envToken == "" {
		return Telegram{}, fmt.Errorf("telegram config: TELEGRAM_BOT_TOKEN is required")
	}

	token, err := domain.NewTelegramToken(envToken)
	if err != nil {
		return Telegram{}, fmt.Errorf("telegram config: %w", err)
	}

	linkTokenTTL, err := getEnvDurationOrDefault("TELEGRAM_LINK_TOKEN_TTL", 15*time.Minute, logger)
	if err != nil {
		return Telegram{}, fmt.Errorf("telegram config: %w", err)
	}

	baseURL := getEnvStringOrDefault("TELEGRAM_API_BASE_URL", "https://api.telegram.org", logger)

	return Telegram{
		Token:        token,
		BaseURL:      baseURL,
		LinkTokenTTL: linkTokenTTL,
	}, nil
}

//nolint:gocritic // Pointer receiver disables formatting
func (c Telegram) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("token", c.Token),
		slog.String("base_url", c.BaseURL),
		slog.Duration("link_token_ttl", c.LinkTokenTTL),
	)
}
