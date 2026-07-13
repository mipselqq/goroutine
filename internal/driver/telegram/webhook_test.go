package telegram_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"goroutine/internal/domain"
	"goroutine/internal/driver/telegram"
	"goroutine/internal/service"
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

func TestWebhookHandler_ServeHTTP(t *testing.T) {
	t.Parallel()

	validToken := testutil.ValidTelegramLinkToken()
	validChatID := testutil.ValidTelegramChatID()

	tests := []struct {
		name          string
		inputBody     any
		setupService  func(s *MockUserService)
		setupNotifier func(n *MockNotifier)
	}{
		{
			name:      "Success",
			inputBody: update("/start "+validToken.RevealSecret(), validChatID.Int64(), "testuser"),
			setupService: func(s *MockUserService) {
				s.LinkTelegramByTokenFunc = func(ctx context.Context, token domain.TelegramLinkToken, chatID domain.TelegramChatID, username domain.TelegramUsername) error {
					return nil
				}
			},
			setupNotifier: func(n *MockNotifier) {
				successMsg := domain.MustTelegramMessage("Successfully linked your account <3")
				n.NotifyFunc = func(ctx context.Context, chatID domain.TelegramChatID, text domain.TelegramMessage) error {
					if chatID != validChatID {
						t.Errorf("got chatID %v, want %v", chatID, validChatID)
					}
					if text != successMsg {
						t.Errorf("got text %q, want success message", text)
					}
					return nil
				}
			},
		},
		{
			name:          "Invalid JSON body",
			inputBody:     []byte("not-json"),
			setupService:  func(s *MockUserService) {},
			setupNotifier: func(n *MockNotifier) {},
		},
		{
			name:          "Non-start message",
			inputBody:     update("/help", validChatID.Int64(), "testuser"),
			setupService:  func(s *MockUserService) {},
			setupNotifier: func(n *MockNotifier) {},
		},
		{
			name:          "Invalid link token in start",
			inputBody:     update("/start not-a-uuid", validChatID.Int64(), "testuser"),
			setupService:  func(s *MockUserService) {},
			setupNotifier: func(n *MockNotifier) {},
		},
		{
			name:          "Invalid chat ID",
			inputBody:     update("/start "+validToken.RevealSecret(), 0, "testuser"),
			setupService:  func(s *MockUserService) {},
			setupNotifier: func(n *MockNotifier) {},
		},
		{
			name:          "Invalid username",
			inputBody:     update("/start "+validToken.RevealSecret(), validChatID.Int64(), ""),
			setupService:  func(s *MockUserService) {},
			setupNotifier: func(n *MockNotifier) {},
		},
		{
			name:      "Token not found",
			inputBody: update("/start "+validToken.RevealSecret(), validChatID.Int64(), "testuser"),
			setupService: func(s *MockUserService) {
				s.LinkTelegramByTokenFunc = func(ctx context.Context, token domain.TelegramLinkToken, chatID domain.TelegramChatID, username domain.TelegramUsername) error {
					return service.ErrTelegramLinkTokenNotFound
				}
			},
			setupNotifier: func(n *MockNotifier) {
				expiredMsg := domain.MustTelegramMessage("This link has expired or is invalid. Please generate a new link in the app.")
				n.NotifyFunc = func(ctx context.Context, chatID domain.TelegramChatID, text domain.TelegramMessage) error {
					if text != expiredMsg {
						t.Errorf("got text %q, want expired message", text)
					}
					return nil
				}
			},
		},
		{
			name:      "User not found",
			inputBody: update("/start "+validToken.RevealSecret(), validChatID.Int64(), "testuser"),
			setupService: func(s *MockUserService) {
				s.LinkTelegramByTokenFunc = func(ctx context.Context, token domain.TelegramLinkToken, chatID domain.TelegramChatID, username domain.TelegramUsername) error {
					return service.ErrUserNotFound
				}
			},
			setupNotifier: func(n *MockNotifier) {
				notFoundMsg := domain.MustTelegramMessage("User account not found.")
				n.NotifyFunc = func(ctx context.Context, chatID domain.TelegramChatID, text domain.TelegramMessage) error {
					if text != notFoundMsg {
						t.Errorf("got text %q, want not found message", text)
					}
					return nil
				}
			},
		},
		{
			name:      "Internal service error",
			inputBody: update("/start "+validToken.RevealSecret(), validChatID.Int64(), "testuser"),
			setupService: func(s *MockUserService) {
				s.LinkTelegramByTokenFunc = func(ctx context.Context, token domain.TelegramLinkToken, chatID domain.TelegramChatID, username domain.TelegramUsername) error {
					return service.ErrInternal
				}
			},
			setupNotifier: func(n *MockNotifier) {
				internalMsg := domain.MustTelegramMessage("Something went wrong. Please try again later.")
				n.NotifyFunc = func(ctx context.Context, chatID domain.TelegramChatID, text domain.TelegramMessage) error {
					if text != internalMsg {
						t.Errorf("got text %q, want generic error message", text)
					}
					return nil
				}
			},
		},
		{
			name:      "Notifier error",
			inputBody: update("/start "+validToken.RevealSecret(), validChatID.Int64(), "testuser"),
			setupService: func(s *MockUserService) {
				s.LinkTelegramByTokenFunc = func(ctx context.Context, token domain.TelegramLinkToken, chatID domain.TelegramChatID, username domain.TelegramUsername) error {
					return nil
				}
			},
			setupNotifier: func(n *MockNotifier) {
				n.NotifyFunc = func(ctx context.Context, chatID domain.TelegramChatID, text domain.TelegramMessage) error {
					return errors.New("network error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, rr := testutil.NewJSONRequestAndRecorder(t, http.MethodPost, "/webhook/telegram", tt.inputBody)

			svc := NewMockUserService(t)
			notifier := NewMockNotifier(t)
			tt.setupService(svc)
			tt.setupNotifier(notifier)

			logger := testutil.NewLogger(t)
			h := telegram.NewWebhookHandler(logger, svc, notifier)
			h.ServeHTTP(rr, req)

			testutil.AssertStatusCode(t, rr, http.StatusOK)
		})
	}
}
