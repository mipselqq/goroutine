package handler_test

import (
	"context"

	"goroutine/internal/domain"
)

type MockAuth struct {
	RegisterFunc func(ctx context.Context, email domain.Email, password domain.Password) error
	LoginFunc    func(ctx context.Context, email domain.Email, password domain.Password) (string, error)
}

func (m *MockAuth) Register(ctx context.Context, email domain.Email, password domain.Password) error {
	return m.RegisterFunc(ctx, email, password)
}

func (m *MockAuth) Login(ctx context.Context, email domain.Email, password domain.Password) (string, error) {
	return m.LoginFunc(ctx, email, password)
}
