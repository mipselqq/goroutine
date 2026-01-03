package service_test

import (
	"context"
	"errors"
	"testing"

	"go-todo/internal/domain"
	"go-todo/internal/repository"
	"go-todo/internal/service"
)

type MockUserRepository struct {
	InsertFunc func(ctx context.Context, email domain.Email, hash string) error
}

func (m *MockUserRepository) Insert(ctx context.Context, email domain.Email, hash string) error {
	if m.InsertFunc != nil {
		return m.InsertFunc(ctx, email, hash)
	}
	return nil
}

func TestAuthService_Register(t *testing.T) {
	t.Parallel()

	// Arrange
	emailStr := "test@example.com"
	passwordStr := "qwerty"

	email, _ := domain.NewEmail(emailStr)
	password, _ := domain.NewPassword(passwordStr)

	tests := []struct {
		name        string
		email       domain.Email
		password    domain.Password
		setupMock   func(r *MockUserRepository)
		expectedErr error
	}{
		{
			name:        "Success",
			email:       email,
			password:    password,
			expectedErr: nil,
			setupMock: func(r *MockUserRepository) {
				r.InsertFunc = func(ctx context.Context, email domain.Email, hash string) error {
					if hash == passwordStr {
						return errors.New("service saved plaintext password!")
					}
					return nil
				}
			},
		},
		{
			name:        "User already exists",
			email:       email,
			password:    password,
			expectedErr: service.ErrUserAlreadyExists,
			setupMock: func(r *MockUserRepository) {
				r.InsertFunc = func(ctx context.Context, email domain.Email, hash string) error {
					return repository.ErrUniqueViolation
				}
			},
		},
		{
			name:        "Internal repository error",
			email:       email,
			password:    password,
			expectedErr: service.ErrInternal,
			setupMock: func(r *MockUserRepository) {
				r.InsertFunc = func(ctx context.Context, email domain.Email, hash string) error {
					return repository.ErrInternal
				}
			},
		},
		{
			name:        "Internal error",
			email:       email,
			password:    password,
			expectedErr: service.ErrInternal,
			setupMock: func(r *MockUserRepository) {
				r.InsertFunc = func(ctx context.Context, email domain.Email, hash string) error {
					return errors.New("some unexpected error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &MockUserRepository{}
			if tt.setupMock != nil {
				tt.setupMock(r)
			}
			s := service.NewAuth(r)

			// Act
			err := s.Register(context.Background(), tt.email, tt.password)

			// Assert
			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}
