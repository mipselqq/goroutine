package telegram

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"goroutine/internal/domain"
)

type APIClient struct {
	botToken domain.TelegramToken
	baseURL  string
	http     *http.Client
	logger   *slog.Logger
}

func NewAPIClient(logger *slog.Logger, baseURL string, botToken domain.TelegramToken) *APIClient {
	return &APIClient{
		botToken: botToken,
		baseURL:  baseURL,
		http:     &http.Client{Timeout: 10 * time.Second},
		logger:   logger,
	}
}

func (c *APIClient) SendMessage(ctx context.Context, chatID int64, text domain.TelegramMessage) error {
	q := url.Values{}
	q.Set("chat_id", fmt.Sprintf("%d", chatID))
	q.Set("text", text.String())

	reqURL := fmt.Sprintf("%s/bot%s/sendMessage?%s", c.baseURL, c.botToken.RevealSecret(), q.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("telegram api: send message: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		c.logger.WarnContext(ctx, "telegram sendMessage failed", slog.String("err", err.Error()), slog.Int64("chat_id", chatID))
		return fmt.Errorf("telegram api: send message: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		c.logger.WarnContext(ctx, "telegram sendMessage non-OK", slog.Int("status", resp.StatusCode), slog.Int64("chat_id", chatID))
		return fmt.Errorf("telegram api: send message: unexpected status %d", resp.StatusCode)
	}

	c.logger.DebugContext(ctx, "telegram sendMessage ok", slog.Int64("chat_id", chatID))
	return nil
}

func (c *APIClient) Notify(ctx context.Context, chatID domain.TelegramChatID, text domain.TelegramMessage) error {
	return c.SendMessage(ctx, chatID.Int64(), text)
}
