package driver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"goroutine/internal/domain"
)

type TelegramClient struct {
	token   domain.TelegramToken
	baseURL string
	http    *http.Client
}

func NewTelegramClient(baseURL string, token domain.TelegramToken) *TelegramClient {
	return &TelegramClient{
		token:   token,
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *TelegramClient) sendMessage(ctx context.Context, chatID int64, text domain.TelegramMessage) error {
	q := url.Values{}
	q.Set("chat_id", fmt.Sprintf("%d", chatID))
	q.Set("text", text.String())

	reqURL := fmt.Sprintf("%s/bot%s/sendMessage?%s", c.baseURL, c.token.RevealSecret(), q.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, http.NoBody)
	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrNetwork, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return telegramResponseError(resp)
	}

	return nil
}

func telegramResponseError(resp *http.Response) *ErrTelegramResponse {
	var parsed struct {
		Description string `json:"description"`
		Parameters  struct {
			RetryAfter int `json:"retry_after"`
		} `json:"parameters"`
	}

	body, err := io.ReadAll(resp.Body)
	if err == nil {
		_ = json.Unmarshal(body, &parsed)
	}

	return newTelegramResponseError(resp.StatusCode, parsed.Description, parsed.Parameters.RetryAfter)
}

func (c *TelegramClient) Notify(ctx context.Context, chatID domain.TelegramChatID, text domain.TelegramMessage) error {
	return c.sendMessage(ctx, chatID.Int64(), text)
}
