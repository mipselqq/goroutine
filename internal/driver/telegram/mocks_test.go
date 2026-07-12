package telegram_test

import (
	"context"
	"testing"

	"goroutine/internal/domain"
	"goroutine/internal/testutil"
)

type MockUserService struct {
	t *testing.T

	LinkTelegramByTokenFunc func(ctx context.Context, token domain.TelegramLinkToken, chatID domain.TelegramChatID, username domain.TelegramUsername) error
}

func NewMockUserService(t *testing.T) *MockUserService {
	return &MockUserService{t: t}
}

func (m *MockUserService) LinkTelegramByToken(ctx context.Context, token domain.TelegramLinkToken, chatID domain.TelegramChatID, username domain.TelegramUsername) error {
	testutil.AssertFuncNotNil(m.t, "UserService.LinkTelegramByTokenFunc", m.LinkTelegramByTokenFunc)
	return m.LinkTelegramByTokenFunc(ctx, token, chatID, username)
}

type MockNotifier struct {
	t *testing.T

	NotifyFunc func(ctx context.Context, chatID domain.TelegramChatID, text domain.TelegramMessage) error
}

func NewMockNotifier(t *testing.T) *MockNotifier {
	return &MockNotifier{t: t}
}

func (m *MockNotifier) Notify(ctx context.Context, chatID domain.TelegramChatID, text domain.TelegramMessage) error {
	testutil.AssertFuncNotNil(m.t, "Notifier.NotifyFunc", m.NotifyFunc)
	return m.NotifyFunc(ctx, chatID, text)
}
