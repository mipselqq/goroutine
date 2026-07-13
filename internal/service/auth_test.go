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
	name          string
	setupUserRepo func(r *MockUserRepository)
	wantErr       error
}

func TestAuth_Register(t *testing.T) {
	t.Parallel()

	tests := []TestCase{
		{
			name:    "Success",
			wantErr: nil,
			setupUserRepo: func(r *MockUserRepository) {
				r.CreateFunc = func(ctx context.Context, email domain.Email, hash domain.PasswordHash) error {
					if hash.RevealSecret() == testutil.ValidPassword().RevealSecret() {
						return errors.New("service saved plaintext password!")
					}
					return nil
				}
			},
		},
		{
			name:    "User already exists",
			wantErr: service.ErrUserAlreadyExists,
			setupUserRepo: func(r *MockUserRepository) {
				r.CreateFunc = func(ctx context.Context, email domain.Email, hash domain.PasswordHash) error {
					return repository.ErrUniqueViolation
				}
			},
		},
		{
			name:    "Internal repository error",
			wantErr: service.ErrInternal,
			setupUserRepo: func(r *MockUserRepository) {
				r.CreateFunc = func(ctx context.Context, email domain.Email, hash domain.PasswordHash) error {
					return repository.ErrInternal
				}
			},
		},
		{
			name:    "Unexpected repository error",
			wantErr: service.ErrInternal,
			setupUserRepo: func(r *MockUserRepository) {
				r.CreateFunc = func(ctx context.Context, email domain.Email, hash domain.PasswordHash) error {
					return errors.New("Super unknown error happened")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := NewMockUserRepository(t)
			tt.setupUserRepo(r)
			s := service.NewAuth(r, testutil.ValidJWTOptions())

			err := s.Register(context.Background(), testutil.ValidEmail(), testutil.ValidPassword())

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestAuth_Login(t *testing.T) {
	t.Parallel()

	tests := []TestCase{
		{
			name: "Success",
			setupUserRepo: func(r *MockUserRepository) {
				r.GetByEmailFunc = func(ctx context.Context, email domain.Email) (domain.User, error) {
					return domain.User{ID: testutil.ValidUserID(), PasswordHash: testutil.ValidPasswordHash()}, nil
				}
			},
			wantErr: nil,
		},
		{
			name:    "User not found",
			wantErr: service.ErrUserNotFound,
			setupUserRepo: func(r *MockUserRepository) {
				r.GetByEmailFunc = func(ctx context.Context, email domain.Email) (domain.User, error) {
					return domain.User{}, repository.ErrRowNotFound
				}
			},
		},
		{
			name:    "Invalid password",
			wantErr: service.ErrInvalidCredentials,
			setupUserRepo: func(r *MockUserRepository) {
				r.GetByEmailFunc = func(ctx context.Context, email domain.Email) (domain.User, error) {
					return domain.User{ID: testutil.ValidUserID(), PasswordHash: testutil.AnotherValidPasswordHash()}, nil
				}
			},
		},
		{
			name:    "Internal repository error",
			wantErr: service.ErrInternal,
			setupUserRepo: func(r *MockUserRepository) {
				r.GetByEmailFunc = func(ctx context.Context, email domain.Email) (domain.User, error) {
					return domain.User{}, repository.ErrInternal
				}
			},
		},
		{
			name:    "Unexpected repository error",
			wantErr: service.ErrInternal,
			setupUserRepo: func(r *MockUserRepository) {
				r.GetByEmailFunc = func(ctx context.Context, email domain.Email) (domain.User, error) {
					return domain.User{}, errors.New("Super unknown error happened")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := NewMockUserRepository(t)
			tt.setupUserRepo(r)
			jwtOpts := testutil.ValidJWTOptions()
			s := service.NewAuth(r, jwtOpts)

			token, err := s.Login(context.Background(), testutil.ValidEmail(), testutil.ValidPassword())

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr == nil {
				parts := strings.Split(token.RevealSecret(), ".")
				if len(parts) != 3 {
					t.Errorf("got %d JWT segments, want 3", len(parts))
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
		name       string
		tokenFunc  func() (domain.AuthToken, error)
		wantUserID domain.UserID
		wantErr    error
	}{
		{
			name: "Valid token",
			tokenFunc: func() (domain.AuthToken, error) {
				return s.CreateToken(testutil.ValidUserID(), jwtOpts.Exp)
			},
			wantUserID: testutil.ValidUserID(),
			wantErr:    nil,
		},
		{
			name: "Invalid token",
			tokenFunc: func() (domain.AuthToken, error) {
				return domain.NewJWTString("invalid.token.here")
			},
			wantErr: service.ErrInvalidToken,
		},
		{
			name: "Different secret",
			tokenFunc: func() (domain.AuthToken, error) {
				claims := jwt.MapClaims{
					"sub": testutil.ValidUserID().String(),
					"exp": time.Now().Add(jwtOpts.Exp).Unix(),
					"iat": time.Now().Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, err := token.SignedString([]byte("wrong_secret"))
				if err != nil {
					return domain.AuthToken{}, err
				}
				return domain.NewJWTString(tokenString)
			},
			wantErr: service.ErrInvalidToken,
		},
		{
			name: "Expired token",
			tokenFunc: func() (domain.AuthToken, error) {
				return s.CreateToken(testutil.ValidUserID(), -time.Hour)
			},
			wantErr: service.ErrTokenExpired,
		},
		{
			name: "Different signing method",
			tokenFunc: func() (domain.AuthToken, error) {
				claims := jwt.MapClaims{
					"sub": testutil.ValidUserID().String(),
					"exp": time.Now().Add(jwtOpts.Exp).Unix(),
					"iat": time.Now().Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
				tokenString, err := token.SignedString([]byte(testutil.ValidJWTSecret().RevealSecret()))
				if err != nil {
					return domain.AuthToken{}, err
				}
				return domain.NewJWTString(tokenString)
			},
			wantErr: service.ErrInvalidSigningMethod,
		},
		{
			name: "Missing sub claim",
			tokenFunc: func() (domain.AuthToken, error) {
				claims := jwt.MapClaims{
					"exp": time.Now().Add(jwtOpts.Exp).Unix(),
					"iat": time.Now().Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, err := token.SignedString([]byte(testutil.ValidJWTSecret().RevealSecret()))
				if err != nil {
					return domain.AuthToken{}, err
				}
				return domain.NewJWTString(tokenString)
			},
			wantErr: service.ErrInvalidToken,
		},
		{
			name: "Sub claim not a string",
			tokenFunc: func() (domain.AuthToken, error) {
				claims := jwt.MapClaims{
					"sub": 12345,
					"exp": time.Now().Add(jwtOpts.Exp).Unix(),
					"iat": time.Now().Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, err := token.SignedString([]byte(testutil.ValidJWTSecret().RevealSecret()))
				if err != nil {
					return domain.AuthToken{}, err
				}
				return domain.NewJWTString(tokenString)
			},
			wantErr: service.ErrInvalidToken,
		},
		{
			name: "Invalid ID in sub",
			tokenFunc: func() (domain.AuthToken, error) {
				claims := jwt.MapClaims{
					"sub": "not-an-id",
					"exp": time.Now().Add(jwtOpts.Exp).Unix(),
					"iat": time.Now().Unix(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, err := token.SignedString([]byte(testutil.ValidJWTSecret().RevealSecret()))
				if err != nil {
					return domain.AuthToken{}, err
				}
				return domain.NewJWTString(tokenString)
			},
			wantErr: service.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			token, err := tt.tokenFunc()
			if err != nil {
				t.Fatalf("tokenFunc() error = %v", err)
			}

			gotUserID, err := s.VerifyToken(context.Background(), token)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr == nil && gotUserID != tt.wantUserID {
				t.Errorf("got userID %q, want %q", gotUserID, tt.wantUserID)
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
		t.Fatalf("CreateToken() error = %v", err)
	}

	parsedToken, err := jwt.Parse(token.RevealSecret(), func(token *jwt.Token) (interface{}, error) {
		return []byte(testutil.ValidJWTSecret().RevealSecret()), nil
	})
	if err != nil {
		t.Fatalf("jwt.Parse() error = %v", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatalf("got claims type %T, want jwt.MapClaims", parsedToken.Claims)
	}

	if claims["sub"] != userID.String() {
		t.Errorf("got sub %v, want %v", claims["sub"], userID.String())
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		t.Fatalf("got exp claim %#v, want float64", claims["exp"])
	}
	wantExp := now.Add(jwtOpts.Exp).Unix()
	if int64(exp) < wantExp-1 || int64(exp) > wantExp+1 {
		t.Errorf("got exp %v, want around %v", int64(exp), wantExp)
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		t.Fatalf("got iat claim %#v, want float64", claims["iat"])
	}
	wantIat := now.Unix()
	if int64(iat) < wantIat-1 || int64(iat) > wantIat+1 {
		t.Errorf("got iat %v, want around %v", int64(iat), wantIat)
	}

	if parsedToken.Method.Alg() != jwt.SigningMethodHS256.Alg() {
		t.Errorf("got alg %v, want %v", parsedToken.Method.Alg(), jwt.SigningMethodHS256.Alg())
	}
}
