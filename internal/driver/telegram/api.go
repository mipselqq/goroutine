package telegram

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"goroutine/internal/domain"
)

type APIClient struct {
	botToken string
	baseURL  string
	http     *http.Client
}

func NewAPIClient(botToken string, httpClient *http.Client) *APIClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &APIClient{
		botToken: botToken,
		baseURL:  "https://api.telegram.org",
		http:     httpClient,
	}
}

func (c *APIClient) SendMessage(ctx context.Context, chatID int64, text string) error {
	q := url.Values{}
	q.Set("chat_id", fmt.Sprintf("%d", chatID))
	q.Set("text", text)

	reqURL := fmt.Sprintf("%s/bot%s/sendMessage?%s", c.baseURL, c.botToken, q.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("telegram api: send message: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("telegram api: send message: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram api: send message: unexpected status %d", resp.StatusCode)
	}

	return nil
}

func (c *APIClient) NotifyLinkSuccess(ctx context.Context, chatID domain.TelegramChatID) error {
	return c.SendMessage(ctx, chatID.Int64(), "Successfully linked your account <3")
}
