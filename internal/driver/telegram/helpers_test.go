package telegram_test

import (
	"context"
	"fmt"

	"goroutine/internal/domain"
)

func AssertFuncNotNil(funcName string, fn any) {
	if fn == nil {
		panic(fmt.Sprintf("%s = nil, want configured mock", funcName))
	}
}

type MockUserService struct {
	LinkTelegramByTokenFunc func(ctx context.Context, token domain.TelegramLinkToken, chatID domain.TelegramChatID, username domain.TelegramUsername) error
}

func (m *MockUserService) LinkTelegramByToken(ctx context.Context, token domain.TelegramLinkToken, chatID domain.TelegramChatID, username domain.TelegramUsername) error {
	AssertFuncNotNil("UserService.LinkTelegramByTokenFunc", m.LinkTelegramByTokenFunc)
	return m.LinkTelegramByTokenFunc(ctx, token, chatID, username)
}

type MockNotifier struct {
	NotifyFunc func(ctx context.Context, chatID domain.TelegramChatID, text domain.TelegramMessage) error
}

func (m *MockNotifier) Notify(ctx context.Context, chatID domain.TelegramChatID, text domain.TelegramMessage) error {
	AssertFuncNotNil("Notifier.NotifyFunc", m.NotifyFunc)
	return m.NotifyFunc(ctx, chatID, text)
}
