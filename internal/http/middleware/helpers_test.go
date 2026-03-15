package middleware_test

import (
	"context"

	"goroutine/internal/domain"
)

type MockAuth struct {
	VerifyTokenFunc func(ctx context.Context, token string) (domain.UserID, error)
}

func (m *MockAuth) VerifyToken(ctx context.Context, token string) (domain.UserID, error) {
	return m.VerifyTokenFunc(ctx, token)
}
