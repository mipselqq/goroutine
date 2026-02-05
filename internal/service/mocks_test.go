package service_test

import (
	"context"

	"goroutine/internal/domain"
)

type MockUserRepository struct {
	InsertFunc                 func(ctx context.Context, email domain.Email, hash string) error
	GetPasswordHashByEmailFunc func(ctx context.Context, email domain.Email) (string, error)
}

func (m *MockUserRepository) Insert(ctx context.Context, email domain.Email, hash string) error {
	return m.InsertFunc(ctx, email, hash)
}

func (m *MockUserRepository) GetPasswordHashByEmail(ctx context.Context, email domain.Email) (string, error) {
	return m.GetPasswordHashByEmailFunc(ctx, email)
}
