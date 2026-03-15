package handler_test

import (
	"context"
	"errors"
	"fmt"
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
	inputBody    string
	setupMock    func(s *MockAuth)
	expectedCode int
	expectedBody string
}

func TestAuth_Register(t *testing.T) {
	t.Parallel()

	email := testutil.ValidEmail()
	password := testutil.ValidPassword()

	tests := []authTestCase{
		{
			name:      "Success",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, email, password),
			setupMock: func(s *MockAuth) {
				s.RegisterFunc = func(ctx context.Context, e domain.Email, p domain.Password) error {
					if e.String() != email.String() {
						t.Errorf("expected email %q, got %q", email, e.String())
					}
					if p.String() != password.String() {
						t.Errorf("expected password %q, got %q", password, p.String())
					}
					return nil
				}
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"status":"ok"}`,
		},
		{
			name:      "Internal error",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, email, password),
			setupMock: func(s *MockAuth) {
				s.RegisterFunc = func(ctx context.Context, email domain.Email, password domain.Password) error {
					return service.ErrInternal
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: fmt.Sprintf(`{"code":"INTERNAL_SERVER_ERROR","message":"Internal server error","timestamp":%q}`, testutil.FixedTime()),
		},
		{
			name:         "Invalid email format",
			inputBody:    fmt.Sprintf(`{"email": %q, "password": %q}`, "invalid-email", password),
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"email","issues":["Invalid email"]}]}`, testutil.FixedTime()),
		},
		{
			name:         "Password too short",
			inputBody:    fmt.Sprintf(`{"email": %q, "password": %q}`, email, "123"),
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"password","issues":["Password is too short"]}]}`, testutil.FixedTime()),
		},
		{
			name:         "Invalid JSON",
			inputBody:    `{"email": "test@example.com", "password": "password"`, // missing closing brace
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"body","issues":["Invalid JSON body"]}]}`, testutil.FixedTime()),
		},
		{
			name:      "User already exists",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, email, password),
			setupMock: func(s *MockAuth) {
				s.RegisterFunc = func(ctx context.Context, email domain.Email, password domain.Password) error {
					return service.ErrUserAlreadyExists
				}
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"INVALID_CREDENTIALS","message":"Invalid login or password","timestamp":%q,"details":[{"field":"email or password","issues":["Invalid credentials"]}]}`, testutil.FixedTime()),
		},
		{
			name:      "Unknown error",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, email, password),
			setupMock: func(s *MockAuth) {
				s.RegisterFunc = func(ctx context.Context, email domain.Email, password domain.Password) error {
					return errors.New("unknown error")
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: fmt.Sprintf(`{"code":"INTERNAL_SERVER_ERROR","message":"Internal server error","timestamp":%q}`, testutil.FixedTime()),
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

			if rr.Code != tt.expectedCode {
				t.Errorf("Expected status %d, got %d", tt.expectedCode, rr.Code)
			}

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
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, email, password),
			setupMock: func(s *MockAuth) {
				s.LoginFunc = func(ctx context.Context, e domain.Email, p domain.Password) (string, error) {
					if e.String() != email.String() {
						t.Errorf("expected email %q, got %q", email, e.String())
					}
					if p.String() != password.String() {
						t.Errorf("expected password %q, got %q", password, p.String())
					}
					return "jwt_token", nil
				}
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"token":"jwt_token"}`,
		},
		{
			name:      "Invalid credentials",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, email, password),
			setupMock: func(s *MockAuth) {
				s.LoginFunc = func(ctx context.Context, email domain.Email, password domain.Password) (string, error) {
					return "", service.ErrInvalidCredentials
				}
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: fmt.Sprintf(`{"code":"INVALID_CREDENTIALS","message":"Invalid login or password","timestamp":%q,"details":[{"field":"email or password","issues":["Invalid"]}]}`, testutil.FixedTime()),
		},
		{
			name:      "User not found",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, email, password),
			setupMock: func(s *MockAuth) {
				s.LoginFunc = func(ctx context.Context, email domain.Email, password domain.Password) (string, error) {
					return "", service.ErrUserNotFound // Enumeration and timing attacks are known, this is fine
				}
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: fmt.Sprintf(`{"code":"USER_NOT_FOUND","message":"User not found","timestamp":%q,"details":[]}`, testutil.FixedTime()),
		},
		{
			name:      "Internal error",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, email, password),
			setupMock: func(s *MockAuth) {
				s.LoginFunc = func(ctx context.Context, email domain.Email, password domain.Password) (string, error) {
					return "", service.ErrInternal
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: fmt.Sprintf(`{"code":"INTERNAL_SERVER_ERROR","message":"Internal server error","timestamp":%q}`, testutil.FixedTime()),
		},
		{
			name:         "Empty email",
			inputBody:    fmt.Sprintf(`{"email": %q, "password": %q}`, "", password),
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"email","issues":["Invalid email"]}]}`, testutil.FixedTime()),
		},
		{
			name:         "Invalid email format",
			inputBody:    fmt.Sprintf(`{"email": %q, "password": %q}`, "invalid-email", password),
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"email","issues":["Invalid email"]}]}`, testutil.FixedTime()),
		},
		{
			name:         "Empty email and password",
			inputBody:    fmt.Sprintf(`{"email": %q, "password": %q}`, "", ""),
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"email","issues":["Invalid email"]},{"field":"password","issues":["Password is too short"]}]}`, testutil.FixedTime()),
		},
		{
			name:         "Empty password",
			inputBody:    fmt.Sprintf(`{"email": %q, "password": %q}`, email, ""),
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"password","issues":["Password is too short"]}]}`, testutil.FixedTime()),
		},
		{
			name:         "Invalid JSON",
			inputBody:    `{"email": "test@example.com"`, // missing password and closing brace
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"body","issues":["Invalid JSON body"]}]}`, testutil.FixedTime()),
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

			if rr.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, rr.Code)
			}

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
		expectedBody string
		setupMock    func(s *MockAuth)
	}{
		{
			name:         "Success",
			context:      context.WithValue(context.Background(), httpschema.ContextKeyUserID, testUserID),
			expectedCode: http.StatusOK,
			expectedBody: fmt.Sprintf(`{"uid":%q}`, testUserID.String()),
		},
		{
			name:         "No user ID",
			context:      context.Background(),
			expectedCode: http.StatusInternalServerError,
			expectedBody: fmt.Sprintf(`{"code":"INTERNAL_SERVER_ERROR","message":"Internal server error","timestamp":%q}`, testutil.FixedTime()),
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

			if rr.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, rr.Code)
			}

			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.expectedBody)
		})
	}
}
