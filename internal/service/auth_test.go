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

	"github.com/golang-jwt/jwt/v5"
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
		JWTSecret:     JWTSecret,
		Exp:           time.Hour,
		SigningMethod: jwt.SigningMethodHS256,
	}
)

func TestAuth_Register(t *testing.T) {
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

func TestAuth_VerifyToken(t *testing.T) {
	t.Parallel()

	s := service.NewAuth(nil, service.JWTOptions{
		JWTSecret:     JWTSecret,
		Exp:           jwtOpts.Exp,
		SigningMethod: jwt.SigningMethodHS256,
	})

	tests := []struct {
		name          string
		tokenFunc     func() (string, error)
		expectedEmail domain.Email
		expectedErr   error
	}{
		{
			name: "Valid token",
			tokenFunc: func() (string, error) {
				return s.CreateToken(email, jwtOpts.Exp)
			},
			expectedEmail: email,
			expectedErr:   nil,
		},
		{
			name: "Invalid token",
			tokenFunc: func() (string, error) {
				return "invalid.token.here", nil
			},
			expectedErr: service.ErrInvalidToken,
		},
		{
			name: "Different secret",
			tokenFunc: func() (string, error) {
				claims := jwt.MapClaims{
					"sub": email.String(),
					"exp": time.Now().Add(jwtOpts.Exp).Unix(),
					"iat": time.Now().Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				return token.SignedString([]byte("wrong_secret"))
			},
			expectedErr: service.ErrInvalidToken,
		},
		{
			name: "Expired token",
			tokenFunc: func() (string, error) {
				return s.CreateToken(email, -time.Hour)
			},
			expectedErr: service.ErrTokenExpired,
		},
		{
			name: "Different sign method",
			tokenFunc: func() (string, error) {
				claims := jwt.MapClaims{
					"sub": email.String(),
					"exp": time.Now().Add(jwtOpts.Exp).Unix(),
					"iat": time.Now().Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
				return token.SignedString([]byte(JWTSecret.RevealSecret()))
			},
			expectedErr: service.ErrInvalidSigningMethod,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			token, err := tt.tokenFunc()
			if err != nil {
				t.Fatalf("token generation failed: %v", err)
			}

			returnedEmail, err := s.VerifyToken(context.Background(), token)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}

			if tt.expectedErr == nil && returnedEmail != tt.expectedEmail {
				t.Errorf("Expected email %v, got %v", tt.expectedEmail, returnedEmail)
			}
		})
	}
}
