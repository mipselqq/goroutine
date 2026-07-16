package handler_test

import (
	"context"
	"net/http"
	"testing"

	"goroutine/internal/domain"
	"goroutine/internal/http/handler"
	"goroutine/internal/testutil"
)

type telegramUpdate struct {
	Message telegramMessage `json:"message"`
}

type telegramMessage struct {
	Text string       `json:"text"`
	Chat telegramChat `json:"chat"`
}

type telegramChat struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

func update(text string, chatID int64, username string) telegramUpdate {
	return telegramUpdate{
		Message: telegramMessage{
			Text: text,
			Chat: telegramChat{ID: chatID, Username: username},
		},
	}
}

func TestTelegramHandler_Webhook(t *testing.T) {
	t.Parallel()

	validToken := testutil.ValidTelegramLinkToken()
	validChatID := testutil.ValidTelegramChatID()
	validUsername := testutil.ValidTelegramUsername()

	tests := []struct {
		name             string
		inputBody        any
		setupUserService func(s *MockUserService)
	}{
		{
			name:      "Success",
			inputBody: update("/start "+validToken.RevealSecret(), validChatID.Int64(), validUsername.String()),
			setupUserService: func(s *MockUserService) {
				s.LinkTelegramByTokenFunc = func(ctx context.Context, token domain.TelegramLinkToken, chatID domain.TelegramChatID, username domain.TelegramUsername) error {
					if token != validToken {
						t.Errorf("got token %v, want %v", token, validToken)
					}
					if chatID != validChatID {
						t.Errorf("got chatID %v, want %v", chatID, validChatID)
					}
					if username != validUsername {
						t.Errorf("got username %v, want %v", username, validUsername)
					}

					return nil
				}
			},
		},
		{
			name:      "Invalid JSON body",
			inputBody: []byte("not-json"),
		},
		{
			name:      "Non-start message",
			inputBody: update("/help", validChatID.Int64(), "testuser"),
		},
		{
			name:      "Invalid link token in start",
			inputBody: update("/start not-a-uuid", validChatID.Int64(), "testuser"),
		},
		{
			name:      "Invalid chat ID",
			inputBody: update("/start "+validToken.RevealSecret(), 0, "testuser"),
		},
		{
			name:      "Invalid username",
			inputBody: update("/start "+validToken.RevealSecret(), validChatID.Int64(), ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, rr := testutil.NewJSONRequestAndRecorder(t, http.MethodPost, "/webhook/telegram", tt.inputBody)

			svc := NewMockUserService(t)

			if tt.setupUserService != nil {
				tt.setupUserService(svc)
			}

			logger := testutil.NewLogger(t)
			h := handler.NewTelegram(logger, svc)
			h.Webhook(rr, req)

			testutil.AssertStatusCode(t, rr, http.StatusOK)
		})
	}
}
