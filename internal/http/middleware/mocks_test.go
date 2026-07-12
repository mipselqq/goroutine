package middleware_test

import (
	"context"
	"testing"

	"goroutine/internal/domain"
	"goroutine/internal/testutil"
)

type MockAuthService struct {
	t *testing.T

	VerifyTokenFunc func(ctx context.Context, token domain.AuthToken) (domain.UserID, error)
}

func NewMockAuthService(t *testing.T) *MockAuthService {
	return &MockAuthService{t: t}
}

func (m *MockAuthService) VerifyToken(ctx context.Context, token domain.AuthToken) (domain.UserID, error) {
	testutil.AssertFuncNotNil(m.t, "AuthService.VerifyTokenFunc", m.VerifyTokenFunc)
	return m.VerifyTokenFunc(ctx, token)
}
