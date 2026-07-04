package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"goroutine/internal/domain"
	"goroutine/internal/http/handler"
	"goroutine/internal/http/httpschema"
	"goroutine/internal/service"
	"goroutine/internal/testutil"
)

type authTestCase struct {
	name             string
	inputBody        any
	setupAuthService func(t *testing.T, s *MockAuthService)
	wantCode         int
	wantBody         any
}

func TestAuth_Register(t *testing.T) {
	t.Parallel()

	email := testutil.ValidEmail()
	password := testutil.ValidPassword()

	tests := []authTestCase{
		{
			name:      "Success",
			inputBody: map[string]string{"email": email.String(), "password": password.RevealSecret()},
			setupAuthService: func(t *testing.T, s *MockAuthService) {
				s.RegisterFunc = func(ctx context.Context, e domain.Email, p domain.UserPassword) error {
					if e != email {
						t.Errorf("got email %q, want %q", e, email)
					}
					if p.RevealSecret() != password.RevealSecret() {
						t.Errorf("got password %q, want %q", p.RevealSecret(), password.RevealSecret())
					}
					return nil
				}
			},
			wantCode: http.StatusOK,
			wantBody: map[string]string{"status": "ok"},
		},
		{
			name:      "Internal error",
			inputBody: map[string]string{"email": email.String(), "password": password.RevealSecret()},
			setupAuthService: func(_ *testing.T, s *MockAuthService) {
				s.RegisterFunc = func(ctx context.Context, email domain.Email, password domain.UserPassword) error {
					return service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: map[string]any{
				"code":      "INTERNAL_SERVER_ERROR",
				"message":   "Internal server error",
				"timestamp": testutil.FixedNowStr(),
			},
		},
		{
			name:      "Invalid email format",
			inputBody: map[string]string{"email": "invalid-email", "password": password.RevealSecret()},
			wantCode:  http.StatusBadRequest,
			wantBody: map[string]any{
				"code":      "VALIDATION_ERROR",
				"message":   "Some fields are invalid",
				"timestamp": testutil.FixedNowStr(),
				"details": []any{
					map[string]any{"field": "email", "issues": []string{"Invalid email"}},
				},
			},
		},
		{
			name:      "Password too short",
			inputBody: map[string]string{"email": email.String(), "password": "123"},
			wantCode:  http.StatusBadRequest,
			wantBody: map[string]any{
				"code":      "VALIDATION_ERROR",
				"message":   "Some fields are invalid",
				"timestamp": testutil.FixedNowStr(),
				"details": []any{
					map[string]any{"field": "password", "issues": []string{"Password is too short"}},
				},
			},
		},
		{
			name:      "Invalid JSON",
			inputBody: json.RawMessage([]byte(`{"email": "test@example.com", "password": "password"`)), // missing closing brace
			wantCode:  http.StatusBadRequest,
			wantBody: map[string]any{
				"code":      "VALIDATION_ERROR",
				"message":   "Some fields are invalid",
				"timestamp": testutil.FixedNowStr(),
				"details": []any{
					map[string]any{"field": "body", "issues": []string{"Invalid JSON body"}},
				},
			},
		},
		{
			name:      "User already exists",
			inputBody: map[string]string{"email": email.String(), "password": password.RevealSecret()},
			setupAuthService: func(_ *testing.T, s *MockAuthService) {
				s.RegisterFunc = func(ctx context.Context, email domain.Email, password domain.UserPassword) error {
					return service.ErrUserAlreadyExists
				}
			},
			wantCode: http.StatusConflict,
			wantBody: userAlreadyExistsError(),
		},
		{
			name:      "Unknown error",
			inputBody: map[string]string{"email": email.String(), "password": password.RevealSecret()},
			setupAuthService: func(_ *testing.T, s *MockAuthService) {
				s.RegisterFunc = func(ctx context.Context, email domain.Email, password domain.UserPassword) error {
					return errors.New("unknown error")
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: map[string]any{
				"code":      "INTERNAL_SERVER_ERROR",
				"message":   "Internal server error",
				"timestamp": testutil.FixedNowStr(),
			},
		},
		{
			name:      "Body too large",
			inputBody: testutil.Big25KBJSON(),
			wantCode:  http.StatusRequestEntityTooLarge,
			wantBody:  payloadTooLargeError(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, rr := testutil.NewJSONRequestAndRecorder(t, http.MethodPost, "/register", tt.inputBody)

			s := MockAuthService{}

			if tt.setupAuthService != nil {
				tt.setupAuthService(t, &s)
			}

			logger := testutil.NewLogger(t)
			h := handler.NewAuth(logger, &s, httpschema.MustNewErrorResponder(logger, testutil.FixedNowStr))
			h.Register(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.wantBody)
		})
	}
}

func TestAuth_Login(t *testing.T) {
	t.Parallel()

	email := testutil.ValidEmail()
	password := testutil.ValidPassword()

	tests := []authTestCase{
		{
			name:      "Success",
			inputBody: map[string]string{"email": email.String(), "password": password.RevealSecret()},
			setupAuthService: func(t *testing.T, s *MockAuthService) {
				s.LoginFunc = func(ctx context.Context, e domain.Email, p domain.UserPassword) (domain.AuthToken, error) {
					if e != email {
						t.Errorf("got email %q, want %q", e, email)
					}
					if p.RevealSecret() != password.RevealSecret() {
						t.Errorf("got password %q, want %q", p.RevealSecret(), password.RevealSecret())
					}
					return domain.NewJWTString("header.payload.signature")
				}
			},
			wantCode: http.StatusOK,
			wantBody: map[string]string{"token": "header.payload.signature"},
		},
		{
			name:      "Invalid credentials",
			inputBody: map[string]string{"email": email.String(), "password": password.RevealSecret()},
			setupAuthService: func(_ *testing.T, s *MockAuthService) {
				s.LoginFunc = func(ctx context.Context, email domain.Email, password domain.UserPassword) (domain.AuthToken, error) {
					return domain.AuthToken{}, service.ErrInvalidCredentials
				}
			},
			wantCode: http.StatusUnauthorized,
			wantBody: map[string]any{
				"code":      "INVALID_CREDENTIALS",
				"message":   "Invalid login or password",
				"timestamp": testutil.FixedNowStr(),
				"details": []any{
					map[string]any{"field": "email or password", "issues": []string{"Invalid"}},
				},
			},
		},
		{
			name:      "User not found",
			inputBody: map[string]string{"email": email.String(), "password": password.RevealSecret()},
			setupAuthService: func(_ *testing.T, s *MockAuthService) {
				s.LoginFunc = func(ctx context.Context, email domain.Email, password domain.UserPassword) (domain.AuthToken, error) {
					return domain.AuthToken{}, service.ErrUserNotFound // Enumeration and timing attacks are known, this is fine
				}
			},
			wantCode: http.StatusUnauthorized,
			wantBody: map[string]any{
				"code":      "USER_NOT_FOUND",
				"message":   "User not found",
				"timestamp": testutil.FixedNowStr(),
				"details":   []any{},
			},
		},
		{
			name:      "Internal error",
			inputBody: map[string]string{"email": email.String(), "password": password.RevealSecret()},
			setupAuthService: func(_ *testing.T, s *MockAuthService) {
				s.LoginFunc = func(ctx context.Context, email domain.Email, password domain.UserPassword) (domain.AuthToken, error) {
					return domain.AuthToken{}, service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: map[string]any{
				"code":      "INTERNAL_SERVER_ERROR",
				"message":   "Internal server error",
				"timestamp": testutil.FixedNowStr(),
			},
		},
		{
			name:      "Empty email",
			inputBody: map[string]string{"email": "", "password": password.RevealSecret()},
			wantCode:  http.StatusBadRequest,
			wantBody: map[string]any{
				"code":      "VALIDATION_ERROR",
				"message":   "Some fields are invalid",
				"timestamp": testutil.FixedNowStr(),
				"details": []any{
					map[string]any{"field": "email", "issues": []string{"Invalid email"}},
				},
			},
		},
		{
			name:      "Invalid email format",
			inputBody: map[string]string{"email": "invalid-email", "password": password.RevealSecret()},
			wantCode:  http.StatusBadRequest,
			wantBody: map[string]any{
				"code":      "VALIDATION_ERROR",
				"message":   "Some fields are invalid",
				"timestamp": testutil.FixedNowStr(),
				"details": []any{
					map[string]any{"field": "email", "issues": []string{"Invalid email"}},
				},
			},
		},
		{
			name:      "Empty email and password",
			inputBody: map[string]any{"email": "", "password": ""},
			wantCode:  http.StatusBadRequest,
			wantBody: map[string]any{
				"code":      "VALIDATION_ERROR",
				"message":   "Some fields are invalid",
				"timestamp": testutil.FixedNowStr(),
				"details": []any{
					map[string]any{"field": "email", "issues": []string{"Invalid email"}},
					map[string]any{"field": "password", "issues": []string{"Password is too short"}},
				},
			},
		},
		{
			name:      "Empty password",
			inputBody: map[string]string{"email": email.String(), "password": ""},
			wantCode:  http.StatusBadRequest,
			wantBody: map[string]any{
				"code":      "VALIDATION_ERROR",
				"message":   "Some fields are invalid",
				"timestamp": testutil.FixedNowStr(),
				"details": []any{
					map[string]any{"field": "password", "issues": []string{"Password is too short"}},
				},
			},
		},
		{
			name:      "Invalid JSON",
			inputBody: json.RawMessage([]byte(`{"email": "test@example.com"`)), // missing password and closing brace
			wantCode:  http.StatusBadRequest,
			wantBody: map[string]any{
				"code":      "VALIDATION_ERROR",
				"message":   "Some fields are invalid",
				"timestamp": testutil.FixedNowStr(),
				"details": []any{
					map[string]any{"field": "body", "issues": []string{"Invalid JSON body"}},
				},
			},
		},
		{
			name:      "Body too large",
			inputBody: testutil.Big25KBJSON(),
			wantCode:  http.StatusRequestEntityTooLarge,
			wantBody:  payloadTooLargeError(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req, rr := testutil.NewJSONRequestAndRecorder(t, http.MethodPost, "/login", tt.inputBody)

			s := &MockAuthService{}

			if tt.setupAuthService != nil {
				tt.setupAuthService(t, s)
			}

			logger := testutil.NewLogger(t)
			h := handler.NewAuth(logger, s, httpschema.MustNewErrorResponder(logger, testutil.FixedNowStr))
			h.Login(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.wantBody)
		})
	}
}

func TestAuth_WhoAmI(t *testing.T) {
	t.Parallel()

	testUserID := testutil.ValidUserID()

	tests := []struct {
		name     string
		context  context.Context
		wantCode int
		wantBody any
	}{
		{
			name:     "Success",
			context:  context.WithValue(context.Background(), httpschema.ContextKeyUserID, testUserID),
			wantCode: http.StatusOK,
			wantBody: map[string]string{"uid": testUserID.String()},
		},
		{
			name:     "No user ID",
			context:  context.Background(),
			wantCode: http.StatusInternalServerError,
			wantBody: map[string]any{
				"code":      "INTERNAL_SERVER_ERROR",
				"message":   "Internal server error",
				"timestamp": testutil.FixedNowStr(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, rr := testutil.NewJSONRequestAndRecorder(t, http.MethodGet, "/whoami", "")
			req = req.WithContext(tt.context)

			logger := testutil.NewLogger(t)
			h := handler.NewAuth(logger, &MockAuthService{}, httpschema.MustNewErrorResponder(logger, testutil.FixedNowStr))
			h.WhoAmI(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.wantBody)
		})
	}
}
