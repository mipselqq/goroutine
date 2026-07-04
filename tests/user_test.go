//go:build e2e

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"goroutine/internal/testutil"
)

func TestUser_TelegramNotifications(t *testing.T) {
	mockTelegram := testutil.NewMockTelegramAPI(t, http.StatusOK)
	defer mockTelegram.Close()

	t.Setenv("TELEGRAM_API_BASE_URL", mockTelegram.URL())

	p := Prelude(t)

	t.Run("Telegram notifications full flow", func(t *testing.T) {
		testutil.TruncateAllTables(t, p.Pool)
		testutil.FlushCurrentRedisDB(t, p.RedisClient)

		ac := CreateUserAndAuthenticateClient(t, p.HTTPClient, p.Server.URL)

		// 1. Get Telegram link token.
		resp := ac.Do(t, http.MethodPost, "/v1/users/me/telegram/link", nil)
		defer func() {
			_ = resp.Body.Close()
		}()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("got status %d, want %d", resp.StatusCode, http.StatusOK)
		}

		var linkBody struct {
			Token string `json:"token"`
		}
		err := json.NewDecoder(resp.Body).Decode(&linkBody)
		if err != nil {
			t.Fatalf("Decode() error = %v", err)
		}

		// 2. Simulate Telegram webhook: user sends /start <token> to the bot.
		webhookBody := map[string]any{
			"message": map[string]any{
				"text": "/start " + linkBody.Token,
				"chat": map[string]any{
					"id":       123456789,
					"username": "testuser",
				},
			},
		}
		bodyBytes, err := json.Marshal(webhookBody)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}
		webhookResp, err := p.HTTPClient.Post(p.Server.URL+"/webhook/telegram", "application/json", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatalf("webhook Post() error = %v", err)
		}
		defer func() {
			_ = webhookResp.Body.Close()
		}()

		if webhookResp.StatusCode != http.StatusOK {
			t.Fatalf("got webhook status %d, want %d", webhookResp.StatusCode, http.StatusOK)
		}

		// 3. Verify the mock Telegram API received the confirmation message.
		if !mockTelegram.Called {
			t.Fatal("mock Telegram API was not called")
		}
		if mockTelegram.LastChatID != 123456789 {
			t.Errorf("got chat_id %d, want %d", mockTelegram.LastChatID, 123456789)
		}
		if !strings.Contains(mockTelegram.LastText, "Successfully linked") {
			t.Errorf("got text %q, want success message", mockTelegram.LastText)
		}
	})
}
