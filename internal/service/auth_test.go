package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"go-todo/internal/domain"
	"go-todo/internal/repository"
	"go-todo/internal/service"
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

type TestCase struct {
	name        string
	setupMock   func(r *MockUserRepository)
	expectedErr error
}

var (
	emailStr            = "test@example.com"
	passwordStr         = "qwerty"
	passwordHash        = "$argon2id$v=19$m=65536,t=1,p=16$kUYJyX3h53cARKnKqFZxvQ$IXz2KOKbyVklgyVmz9ebJ1ffOgmcyMpn/GTUWsep5lk"
	email, _            = domain.NewEmail(emailStr)
	password, _         = domain.NewPassword(passwordStr)
	anotherPasswordHash = "$argon2id$v=19$m=65536,t=3,p=4$bm90LXF3ZXJ0eQ$fSowp1Rof0fXhF+rXv2f6w"
	JWTSecret           = "secret"
)

func TestAuthService_Register(t *testing.T) {
	t.Parallel()

	tests := []TestCase{
		{
			name:        "Success",
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
			expectedErr: service.ErrUserAlreadyExists,
			setupMock: func(r *MockUserRepository) {
				r.InsertFunc = func(ctx context.Context, email domain.Email, hash string) error {
					return repository.ErrUniqueViolation
				}
			},
		},
		{
			name:        "Internal repository error",
			expectedErr: service.ErrInternal,
			setupMock: func(r *MockUserRepository) {
				r.InsertFunc = func(ctx context.Context, email domain.Email, hash string) error {
					return repository.ErrInternal
				}
			},
		},
		{
			name:        "Unexpected repository error",
			expectedErr: service.ErrInternal,
			setupMock: func(r *MockUserRepository) {
				r.InsertFunc = func(ctx context.Context, email domain.Email, hash string) error {
					return errors.New("Super unknown error happened")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &MockUserRepository{}
			tt.setupMock(r)
			s := service.NewAuth(r, JWTSecret)

			err := s.Register(context.Background(), email, password)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestAuth_Login(t *testing.T) {
	t.Parallel()

	tests := []TestCase{
		{
			name: "Success",
			setupMock: func(r *MockUserRepository) {
				r.GetPasswordHashByEmailFunc = func(ctx context.Context, email domain.Email) (string, error) {
					return passwordHash, nil
				}
			},
			expectedErr: nil,
		},
		{
			name:        "User not found",
			expectedErr: service.ErrUserNotFound,
			setupMock: func(r *MockUserRepository) {
				r.GetPasswordHashByEmailFunc = func(ctx context.Context, email domain.Email) (string, error) {
					return "", repository.ErrRowNotFound
				}
			},
		},
		{
			name:        "Invalid password",
			expectedErr: service.ErrInvalidCredentials,
			setupMock: func(r *MockUserRepository) {
				r.GetPasswordHashByEmailFunc = func(ctx context.Context, email domain.Email) (string, error) {
					return anotherPasswordHash, nil
				}
			},
		},
		{
			name:        "Internal repository error",
			expectedErr: service.ErrInternal,
			setupMock: func(r *MockUserRepository) {
				r.GetPasswordHashByEmailFunc = func(ctx context.Context, email domain.Email) (string, error) {
					return "", repository.ErrInternal
				}
			},
		},
		{
			name:        "Unexpected repository error",
			expectedErr: service.ErrInternal,
			setupMock: func(r *MockUserRepository) {
				r.GetPasswordHashByEmailFunc = func(ctx context.Context, email domain.Email) (string, error) {
					return "", errors.New("Super unknown error happened")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &MockUserRepository{}
			tt.setupMock(r)
			s := service.NewAuth(r, JWTSecret)

			token, err := s.Login(context.Background(), email, password)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}

			if tt.expectedErr == nil {
				parts := strings.Split(token, ".")
				if len(parts) != 3 {
					t.Errorf("Expected JWT token, got %s", token)
				}
			}
		})
	}
}
