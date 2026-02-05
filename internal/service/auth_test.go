package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/secrecy"
	"goroutine/internal/service"
)

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
	JWTSecret           = secrecy.SecretString("secret")
	jwtOpts             = service.JWTOptions{
		JWTSecret: JWTSecret,
		Exp:       time.Hour,
	}
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
			s := service.NewAuth(r, jwtOpts)

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
			s := service.NewAuth(r, jwtOpts)

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

func TestAuthService_VerifyToken(t *testing.T) {
	t.Parallel()

	r := &MockUserRepository{
		GetPasswordHashByEmailFunc: func(ctx context.Context, email domain.Email) (string, error) {
			return passwordHash, nil
		},
	}
	s := service.NewAuth(r, jwtOpts)

	t.Run("Valid token", func(t *testing.T) {
		t.Parallel()

		token, err := service.CreateToken(email, JWTSecret.RevealSecret(), jwtOpts.Exp)
		if err != nil {
			t.Fatalf("CreateToken failed: %v", err)
		}

		returnedEmail, err := s.VerifyToken(context.Background(), token)
		if err != nil {
			t.Fatalf("VerifyToken failed: %v", err)
		}

		if returnedEmail != email {
			t.Errorf("Expected email %v, got %v", email, returnedEmail)
		}
	})

	t.Run("Invalid token", func(t *testing.T) {
		t.Parallel()

		_, err := s.VerifyToken(context.Background(), "invalid.token.here")
		if !errors.Is(err, service.ErrInvalidToken) {
			t.Errorf("Expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("Invalid secret", func(t *testing.T) {
		t.Parallel()

		token, err := service.CreateToken(email, JWTSecret.RevealSecret(), jwtOpts.Exp)
		if err != nil {
			t.Fatalf("CreateToken failed: %v", err)
		}

		invalidSecretService := service.NewAuth(nil, service.JWTOptions{JWTSecret: "wrong_secret", Exp: time.Hour})
		_, err = invalidSecretService.VerifyToken(context.Background(), token)
		if !errors.Is(err, service.ErrInvalidToken) {
			t.Errorf("Expected ErrInvalidToken, got %v", err)
		}
	})

	t.Run("Expired token", func(t *testing.T) {
		t.Parallel()

		token, err := service.CreateToken(email, JWTSecret.RevealSecret(), time.Second)
		if err != nil {
			t.Fatalf("CreateToken failed: %v", err)
		}

		time.Sleep(2 * time.Second)

		_, err = s.VerifyToken(context.Background(), token)
		if !errors.Is(err, service.ErrTokenExpired) {
			t.Errorf("Expected ErrTokenExpired, got %v", err)
		}
	})
}
