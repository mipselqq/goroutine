package service_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/service"
	"goroutine/internal/testutil"

	"github.com/golang-jwt/jwt/v5"
)

type TestCase struct {
	name        string
	setupMock   func(r *MockUserRepository)
	expectedErr error
}

func TestAuth_Register(t *testing.T) {
	t.Parallel()

	tests := []TestCase{
		{
			name:        "Success",
			expectedErr: nil,
			setupMock: func(r *MockUserRepository) {
				r.InsertFunc = func(ctx context.Context, email domain.Email, hash string) error {
					if hash == testutil.ValidPassword().String() {
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
			s := service.NewAuth(r, testutil.ValidJWTOptions())

			err := s.Register(context.Background(), testutil.ValidEmail(), testutil.ValidPassword())

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
				r.GetByEmailFunc = func(ctx context.Context, email domain.Email) (domain.UserID, string, error) {
					return testutil.ValidUserID(), testutil.ValidPasswordHash(), nil
				}
			},
			expectedErr: nil,
		},
		{
			name:        "User not found",
			expectedErr: service.ErrUserNotFound,
			setupMock: func(r *MockUserRepository) {
				r.GetByEmailFunc = func(ctx context.Context, email domain.Email) (domain.UserID, string, error) {
					return domain.UserID{}, "", repository.ErrRowNotFound
				}
			},
		},
		{
			name:        "Invalid password",
			expectedErr: service.ErrInvalidCredentials,
			setupMock: func(r *MockUserRepository) {
				r.GetByEmailFunc = func(ctx context.Context, email domain.Email) (domain.UserID, string, error) {
					return testutil.ValidUserID(), testutil.AnotherValidPasswordHash(), nil
				}
			},
		},
		{
			name:        "Internal repository error",
			expectedErr: service.ErrInternal,
			setupMock: func(r *MockUserRepository) {
				r.GetByEmailFunc = func(ctx context.Context, email domain.Email) (domain.UserID, string, error) {
					return domain.UserID{}, "", repository.ErrInternal
				}
			},
		},
		{
			name:        "Unexpected repository error",
			expectedErr: service.ErrInternal,
			setupMock: func(r *MockUserRepository) {
				r.GetByEmailFunc = func(ctx context.Context, email domain.Email) (domain.UserID, string, error) {
					return domain.UserID{}, "", errors.New("Super unknown error happened")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &MockUserRepository{}
			tt.setupMock(r)
			jwtOpts := testutil.ValidJWTOptions()
			s := service.NewAuth(r, jwtOpts)

			token, err := s.Login(context.Background(), testutil.ValidEmail(), testutil.ValidPassword())

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}

			if tt.expectedErr == nil {
				parts := strings.Split(token, ".")
				if len(parts) != 3 {
					t.Errorf("Expected JWT token, got %q", token)
				}
			}
		})
	}
}

func TestAuth_VerifyToken(t *testing.T) {
	t.Parallel()

	jwtOpts := testutil.ValidJWTOptions()
	s := service.NewAuth(nil, jwtOpts)

	tests := []struct {
		name           string
		tokenFunc      func() (string, error)
		expectedUserID domain.UserID
		expectedErr    error
	}{
		{
			name: "Valid token",
			tokenFunc: func() (string, error) {
				return s.CreateToken(testutil.ValidUserID(), jwtOpts.Exp)
			},
			expectedUserID: testutil.ValidUserID(),
			expectedErr:    nil,
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
					"sub": testutil.ValidUserID().String(),
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
				return s.CreateToken(testutil.ValidUserID(), -time.Hour)
			},
			expectedErr: service.ErrTokenExpired,
		},
		{
			name: "Different signing method",
			tokenFunc: func() (string, error) {
				claims := jwt.MapClaims{
					"sub": testutil.ValidUserID().String(),
					"exp": time.Now().Add(jwtOpts.Exp).Unix(),
					"iat": time.Now().Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
				return token.SignedString([]byte(testutil.ValidJWTSecret().RevealSecret()))
			},
			expectedErr: service.ErrInvalidSigningMethod,
		},
		{
			name: "Missing sub claim",
			tokenFunc: func() (string, error) {
				claims := jwt.MapClaims{
					"exp": time.Now().Add(jwtOpts.Exp).Unix(),
					"iat": time.Now().Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				return token.SignedString([]byte(testutil.ValidJWTSecret().RevealSecret()))
			},
			expectedErr: service.ErrInvalidToken,
		},
		{
			name: "Sub claim not a string",
			tokenFunc: func() (string, error) {
				claims := jwt.MapClaims{
					"sub": 12345,
					"exp": time.Now().Add(jwtOpts.Exp).Unix(),
					"iat": time.Now().Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				return token.SignedString([]byte(testutil.ValidJWTSecret().RevealSecret()))
			},
			expectedErr: service.ErrInvalidToken,
		},
		{
			name: "Invalid ID in sub",
			tokenFunc: func() (string, error) {
				claims := jwt.MapClaims{
					"sub": "not-an-id",
					"exp": time.Now().Add(jwtOpts.Exp).Unix(),
					"iat": time.Now().Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				return token.SignedString([]byte(testutil.ValidJWTSecret().RevealSecret()))
			},
			expectedErr: service.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			token, err := tt.tokenFunc()
			if err != nil {
				t.Fatalf("token generation failed: %v", err)
			}

			returnedUserID, err := s.VerifyToken(context.Background(), token)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("Expected error %v, got %v", tt.expectedErr, err)
			}

			if tt.expectedErr == nil && returnedUserID != tt.expectedUserID {
				t.Errorf("Expected userID %q, got %q", tt.expectedUserID, returnedUserID)
			}
		})
	}
}

func TestAuth_CreateToken(t *testing.T) {
	t.Parallel()

	jwtOpts := testutil.ValidJWTOptions()
	s := service.NewAuth(nil, jwtOpts)
	now := time.Now()
	userID := testutil.ValidUserID()
	token, err := s.CreateToken(userID, jwtOpts.Exp)
	if err != nil {
		t.Fatalf("token creation failed: %v", err)
	}

	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(testutil.ValidJWTSecret().RevealSecret()), nil
	})
	if err != nil {
		t.Fatalf("token parsing failed: %v", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("claims are not map %v", claims)
	}

	if claims["sub"] != userID.String() {
		t.Errorf("Expected sub %v, got %v", userID, claims["sub"])
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		t.Fatalf("exp claim is missing or not a number")
	}
	expectedExp := now.Add(jwtOpts.Exp).Unix()
	if int64(exp) < expectedExp-1 || int64(exp) > expectedExp+1 {
		t.Errorf("Expected exp around %v, got %v", expectedExp, int64(exp))
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		t.Fatalf("iat claim is missing or not a number")
	}
	expectedIat := now.Unix()
	if int64(iat) < expectedIat-1 || int64(iat) > expectedIat+1 {
		t.Errorf("Expected iat around %v, got %v", expectedIat, int64(iat))
	}

	if parsedToken.Method.Alg() != jwt.SigningMethodHS256.Alg() {
		t.Errorf("Expected alg %v, got %v", jwt.SigningMethodHS256.Alg(), parsedToken.Method.Alg())
	}
}
