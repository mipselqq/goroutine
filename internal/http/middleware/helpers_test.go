package middleware_test

import (
	"context"
	"fmt"

	"goroutine/internal/domain"
)

type MockAuthService struct {
	VerifyTokenFunc func(ctx context.Context, token string) (domain.UserID, error)
}

func AssertFuncNotNil(funcName string, fn any) {
	if fn == nil {
		panic(fmt.Sprintf("%s = nil, want configured mock", funcName))
	}
}

func (m *MockAuthService) VerifyToken(ctx context.Context, token string) (domain.UserID, error) {
	AssertFuncNotNil("AuthService.VerifyTokenFunc", m.VerifyTokenFunc)
	return m.VerifyTokenFunc(ctx, token)
}
