//go:build e2e

package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"goroutine/internal/testutil"
)

func TestNotification_HappyPath(t *testing.T) {
	mockTelegram := testutil.NewMockTelegramAPI(t, http.StatusOK)
	defer mockTelegram.Close()

	t.Setenv("TELEGRAM_API_BASE_URL", mockTelegram.URL())

	p := prelude(t)

	testutil.TruncateAllTables(t, p.Pool)
	testutil.FlushRedisDB(t, p.RedisClient)

	startBackgroundNotificationsWorker(t, p.Pool)

	ac := createUserAndAuthenticateClient(t, p.HTTPClient, p.Server.URL)

	// 1. Get Telegram link token
	linkResp := ac.Do(t, http.MethodPost, "/v1/users/me/telegram/link", nil)
	defer func() {
		_ = linkResp.Body.Close()
	}()

	if linkResp.StatusCode != http.StatusOK {
		t.Fatalf("got status %d, want %d", linkResp.StatusCode, http.StatusOK)
	}

	var linkBody struct {
		Token string `json:"token"`
	}
	err := json.NewDecoder(linkResp.Body).Decode(&linkBody)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	// 2. Simulate Telegram webhook: user sends /start <token> to the bot
	webhookBody := map[string]any{
		"message": map[string]any{
			"text": "/start " + linkBody.Token,
			"chat": map[string]any{
				"id":       123456789,
				"username": "testuser",
			},
		},
	}
	body, err := json.Marshal(webhookBody)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	webhookResp, err := p.HTTPClient.Post(p.Server.URL+"/webhook/telegram", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("webhook Post() error = %v", err)
	}
	defer func() {
		_ = webhookResp.Body.Close()
	}()

	if webhookResp.StatusCode != http.StatusOK {
		t.Fatalf("got webhook status %d, want %d", webhookResp.StatusCode, http.StatusOK)
	}

	mockTelegram.Called = false

	// 3. Create a board.
	createResp := ac.Do(t, http.MethodPost, "/v1/boards", map[string]string{
		"name":        testutil.ValidBoardName().String(),
		"description": testutil.ValidBoardDescription().String(),
	})
	defer func() {
		_ = createResp.Body.Close()
	}()

	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("got board create status %d, want %d", createResp.StatusCode, http.StatusCreated)
	}

	// 4. Wait for the notification worker to send the board-created event
	time.Sleep(2 * time.Second)

	if !mockTelegram.Called {
		t.Fatal("got no notification, want in 2 seconds after board creation")
	}
}
