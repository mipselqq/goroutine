package middleware_test

import "context"

type MockAuth struct {
	VerifyTokenFunc func(ctx context.Context, token string) (int64, error)
}

func (m *MockAuth) VerifyToken(ctx context.Context, token string) (int64, error) {
	return m.VerifyTokenFunc(ctx, token)
}
