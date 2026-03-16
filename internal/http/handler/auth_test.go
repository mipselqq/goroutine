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
	name         string
	inputBody    any
	setupMock    func(s *MockAuth)
	expectedCode int
	expectedBody any
}

func TestAuth_Register(t *testing.T) {
	t.Parallel()

	email := testutil.ValidEmail()
	password := testutil.ValidPassword()

	tests := []authTestCase{
		{
			name:      "Success",
			inputBody: map[string]string{"email": email.String(), "password": password.String()},
			setupMock: func(s *MockAuth) {
				s.RegisterFunc = func(ctx context.Context, e domain.Email, p domain.UserPassword) error {
					if e != email {
						t.Errorf("expected email %q, got %q", email, e)
					}
					if p != password {
						t.Errorf("expected password %q, got %q", password, p)
					}
					return nil
				}
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]string{"status": "ok"},
		},
		{
			name:      "Internal error",
			inputBody: map[string]string{"email": email.String(), "password": password.String()},
			setupMock: func(s *MockAuth) {
				s.RegisterFunc = func(ctx context.Context, email domain.Email, password domain.UserPassword) error {
					return service.ErrInternal
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: map[string]any{
				"code":      "INTERNAL_SERVER_ERROR",
				"message":   "Internal server error",
				"timestamp": testutil.FixedTime(),
			},
		},
		{
			name:         "Invalid email format",
			inputBody:    map[string]string{"email": "invalid-email", "password": password.String()},
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]any{
				"code":      "VALIDATION_ERROR",
				"message":   "Some fields are invalid",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "email", "issues": []string{"Invalid email"}},
				},
			},
		},
		{
			name:         "Password too short",
			inputBody:    map[string]string{"email": email.String(), "password": "123"},
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]any{
				"code":      "VALIDATION_ERROR",
				"message":   "Some fields are invalid",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "password", "issues": []string{"Password is too short"}},
				},
			},
		},
		{
			name:         "Invalid JSON",
			inputBody:    json.RawMessage([]byte(`{"email": "test@example.com", "password": "password"`)), // missing closing brace
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]any{
				"code":      "VALIDATION_ERROR",
				"message":   "Some fields are invalid",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "body", "issues": []string{"Invalid JSON body"}},
				},
			},
		},
		{
			name:      "User already exists",
			inputBody: map[string]string{"email": email.String(), "password": password.String()},
			setupMock: func(s *MockAuth) {
				s.RegisterFunc = func(ctx context.Context, email domain.Email, password domain.UserPassword) error {
					return service.ErrUserAlreadyExists
				}
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]any{
				"code":      "INVALID_CREDENTIALS",
				"message":   "Invalid login or password",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "email or password", "issues": []string{"Invalid credentials"}},
				},
			},
		},
		{
			name:      "Unknown error",
			inputBody: map[string]string{"email": email.String(), "password": password.String()},
			setupMock: func(s *MockAuth) {
				s.RegisterFunc = func(ctx context.Context, email domain.Email, password domain.UserPassword) error {
					return errors.New("unknown error")
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: map[string]any{
				"code":      "INTERNAL_SERVER_ERROR",
				"message":   "Internal server error",
				"timestamp": testutil.FixedTime(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, rr := testutil.NewJSONRequestAndRecorder(t, http.MethodPost, "/register", tt.inputBody)

			s := MockAuth{}

			if tt.setupMock != nil {
				tt.setupMock(&s)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewAuth(logger, &s, httpschema.MustNewErrorResponder(logger, testutil.FixedTime))
			h.Register(rr, req)

			testutil.AssertStatusCode(t, rr, tt.expectedCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.expectedBody)
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
			inputBody: map[string]string{"email": email.String(), "password": password.String()},
			setupMock: func(s *MockAuth) {
				s.LoginFunc = func(ctx context.Context, e domain.Email, p domain.UserPassword) (string, error) {
					if e != email {
						t.Errorf("expected email %q, got %q", email, e)
					}
					if p != password {
						t.Errorf("expected password %q, got %q", password, p)
					}
					return "jwt_token", nil
				}
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]string{"token": "jwt_token"},
		},
		{
			name:      "Invalid credentials",
			inputBody: map[string]string{"email": email.String(), "password": password.String()},
			setupMock: func(s *MockAuth) {
				s.LoginFunc = func(ctx context.Context, email domain.Email, password domain.UserPassword) (string, error) {
					return "", service.ErrInvalidCredentials
				}
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: map[string]any{
				"code":      "INVALID_CREDENTIALS",
				"message":   "Invalid login or password",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "email or password", "issues": []string{"Invalid"}},
				},
			},
		},
		{
			name:      "User not found",
			inputBody: map[string]string{"email": email.String(), "password": password.String()},
			setupMock: func(s *MockAuth) {
				s.LoginFunc = func(ctx context.Context, email domain.Email, password domain.UserPassword) (string, error) {
					return "", service.ErrUserNotFound // Enumeration and timing attacks are known, this is fine
				}
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: map[string]any{
				"code":      "USER_NOT_FOUND",
				"message":   "User not found",
				"timestamp": testutil.FixedTime(),
				"details":   []any{},
			},
		},
		{
			name:      "Internal error",
			inputBody: map[string]string{"email": email.String(), "password": password.String()},
			setupMock: func(s *MockAuth) {
				s.LoginFunc = func(ctx context.Context, email domain.Email, password domain.UserPassword) (string, error) {
					return "", service.ErrInternal
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: map[string]any{
				"code":      "INTERNAL_SERVER_ERROR",
				"message":   "Internal server error",
				"timestamp": testutil.FixedTime(),
			},
		},
		{
			name:         "Empty email",
			inputBody:    map[string]string{"email": "", "password": password.String()},
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]any{
				"code":      "VALIDATION_ERROR",
				"message":   "Some fields are invalid",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "email", "issues": []string{"Invalid email"}},
				},
			},
		},
		{
			name:         "Invalid email format",
			inputBody:    map[string]string{"email": "invalid-email", "password": password.String()},
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]any{
				"code":      "VALIDATION_ERROR",
				"message":   "Some fields are invalid",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "email", "issues": []string{"Invalid email"}},
				},
			},
		},
		{
			name:         "Empty email and password",
			inputBody:    map[string]any{"email": "", "password": ""},
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]any{
				"code":      "VALIDATION_ERROR",
				"message":   "Some fields are invalid",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "email", "issues": []string{"Invalid email"}},
					map[string]any{"field": "password", "issues": []string{"Password is too short"}},
				},
			},
		},
		{
			name:         "Empty password",
			inputBody:    map[string]string{"email": email.String(), "password": ""},
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]any{
				"code":      "VALIDATION_ERROR",
				"message":   "Some fields are invalid",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "password", "issues": []string{"Password is too short"}},
				},
			},
		},
		{
			name:         "Invalid JSON",
			inputBody:    json.RawMessage([]byte(`{"email": "test@example.com"`)), // missing password and closing brace
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]any{
				"code":      "VALIDATION_ERROR",
				"message":   "Some fields are invalid",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "body", "issues": []string{"Invalid JSON body"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			req, rr := testutil.NewJSONRequestAndRecorder(t, http.MethodPost, "/login", tt.inputBody)

			s := &MockAuth{}

			if tt.setupMock != nil {
				tt.setupMock(s)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewAuth(logger, s, httpschema.MustNewErrorResponder(logger, testutil.FixedTime))
			h.Login(rr, req)

			testutil.AssertStatusCode(t, rr, tt.expectedCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.expectedBody)
		})
	}
}

func TestAuth_WhoAmI(t *testing.T) {
	testUserID := testutil.ValidUserID()

	tests := []struct {
		name         string
		context      context.Context
		expectedCode int
		expectedBody any
		setupMock    func(s *MockAuth)
	}{
		{
			name:         "Success",
			context:      context.WithValue(context.Background(), httpschema.ContextKeyUserID, testUserID),
			expectedCode: http.StatusOK,
			expectedBody: map[string]string{"uid": testUserID.String()},
		},
		{
			name:         "No user ID",
			context:      context.Background(),
			expectedCode: http.StatusInternalServerError,
			expectedBody: map[string]any{
				"code":      "INTERNAL_SERVER_ERROR",
				"message":   "Internal server error",
				"timestamp": testutil.FixedTime(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, rr := testutil.NewJSONRequestAndRecorder(t, http.MethodGet, "/whoami", "")
			req = req.WithContext(tt.context)

			logger := testutil.NewTestLogger(t)
			h := handler.NewAuth(logger, &MockAuth{}, httpschema.MustNewErrorResponder(logger, testutil.FixedTime))
			h.WhoAmI(rr, req)

			testutil.AssertStatusCode(t, rr, tt.expectedCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.expectedBody)
		})
	}
}
