package telegram

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"goroutine/internal/domain"
)

type APIClient struct {
	botToken domain.TelegramToken
	baseURL  string
	http     *http.Client
}

func NewAPIClient(baseURL string, botToken domain.TelegramToken) *APIClient {
	return &APIClient{
		botToken: botToken,
		baseURL:  baseURL,
		http:     &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *APIClient) SendMessage(ctx context.Context, chatID int64, text domain.TelegramMessage) error {
	q := url.Values{}
	q.Set("chat_id", fmt.Sprintf("%d", chatID))
	q.Set("text", text.String())

	reqURL := fmt.Sprintf("%s/bot%s/sendMessage?%s", c.baseURL, c.botToken.RevealSecret(), q.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, http.NoBody)
	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return nil
}

func (c *APIClient) Notify(ctx context.Context, chatID domain.TelegramChatID, text domain.TelegramMessage) error {
	return c.SendMessage(ctx, chatID.Int64(), text)
}
