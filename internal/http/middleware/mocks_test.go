package middleware_test

import (
	"context"

	"goroutine/internal/domain"
)

const FixedTime string = "2026-01-01T00:00:00Z"

func MockTime() string { return FixedTime }

type MockAuth struct {
	VerifyTokenFunc func(ctx context.Context, token string) (domain.UserID, error)
}

func (m *MockAuth) VerifyToken(ctx context.Context, token string) (domain.UserID, error) {
	return m.VerifyTokenFunc(ctx, token)
}
