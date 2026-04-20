package middleware_test

import (
	"context"

	"goroutine/internal/domain"
)

type MockAuthService struct {
	VerifyTokenFunc func(ctx context.Context, token string) (domain.UserID, error)
}

func (m *MockAuthService) VerifyToken(ctx context.Context, token string) (domain.UserID, error) {
	return m.VerifyTokenFunc(ctx, token)
}
