package handler_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/http/httptest"
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

	tests := []authTestCase{
		{
			name:      "Success",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, Email, Password),
			setupMock: func(s *MockAuth) {
				s.RegisterFunc = func(ctx context.Context, e domain.Email, p domain.Password) error {
					if e.String() != Email {
						t.Errorf("expected email %q, got %q", Email, e.String())
					}
					if p.String() != Password {
						t.Errorf("expected password %q, got %q", Password, p.String())
					}
					return nil
				}
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"status":"ok"}`,
		},
		{
			name:      "Internal error",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, Email, Password),
			setupMock: func(s *MockAuth) {
				s.RegisterFunc = func(ctx context.Context, email domain.Email, password domain.Password) error {
					return service.ErrInternal
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: fmt.Sprintf(`{"code":"INTERNAL_SERVER_ERROR","message":"Internal server error","timestamp":%q}`, FixedTime),
		},
		{
			name:         "Invalid email format",
			inputBody:    fmt.Sprintf(`{"email": %q, "password": %q}`, "invalid-email", Password),
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"email","issues":["Invalid email"]}]}`, FixedTime),
		},
		{
			name:         "Password too short",
			inputBody:    fmt.Sprintf(`{"email": %q, "password": %q}`, Email, "123"),
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"password","issues":["Password is too short"]}]}`, FixedTime),
		},
		{
			name:         "Invalid JSON",
			inputBody:    `{"email": "test@example.com", "password": "password"`, // missing closing brace
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"body","issues":["Invalid JSON body"]}]}`, FixedTime),
		},
		{
			name:      "User already exists",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, Email, Password),
			setupMock: func(s *MockAuth) {
				s.RegisterFunc = func(ctx context.Context, email domain.Email, password domain.Password) error {
					return service.ErrUserAlreadyExists
				}
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"INVALID_CREDENTIALS","message":"Invalid login or password","timestamp":%q,"details":[{"field":"email or password","issues":["Invalid credentials"]}]}`, FixedTime),
		},
		{
			name:      "Unknown error",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, Email, Password),
			setupMock: func(s *MockAuth) {
				s.RegisterFunc = func(ctx context.Context, email domain.Email, password domain.Password) error {
					return errors.New("unknown error")
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: fmt.Sprintf(`{"code":"INTERNAL_SERVER_ERROR","message":"Internal server error","timestamp":%q}`, FixedTime),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := testutil.NewTestLogger(t)

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte(tt.inputBody)))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			s := MockAuth{}

			if tt.setupMock != nil {
				tt.setupMock(&s)
			}

			h := handler.NewAuth(logger, &s, httpschema.MustNewErrorResponder(logger, MockTime))
			h.Register(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("Expected status %d, got %d", tt.expectedCode, rr.Code)
			}

			contentType := rr.Header().Get("Content-Type")
			mediaType, _, err := mime.ParseMediaType(contentType)
			if err != nil {
				t.Fatalf("Failed to parse MIME %q", contentType)
			}
			if mediaType != ExpectedMime {
				t.Errorf("Expected %q, got %q", ExpectedMime, mediaType)
			}

			testutil.AssertResponseBody(t, rr, tt.expectedBody)
		})
	}
}

func TestAuth_Login(t *testing.T) {
	t.Parallel()

	tests := []authTestCase{
		{
			name:      "Success",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, Email, Password),
			setupMock: func(s *MockAuth) {
				s.LoginFunc = func(ctx context.Context, e domain.Email, p domain.Password) (string, error) {
					if e.String() != Email {
						t.Errorf("expected email %q, got %q", Email, e.String())
					}
					if p.String() != Password {
						t.Errorf("expected password %q, got %q", Password, p.String())
					}
					return "jwt_token", nil
				}
			},
			expectedCode: http.StatusOK,
			expectedBody: `{"token":"jwt_token"}`,
		},
		{
			name:      "Invalid credentials",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, Email, Password),
			setupMock: func(s *MockAuth) {
				s.LoginFunc = func(ctx context.Context, email domain.Email, password domain.Password) (string, error) {
					return "", service.ErrInvalidCredentials
				}
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: fmt.Sprintf(`{"code":"INVALID_CREDENTIALS","message":"Invalid login or password","timestamp":%q,"details":[{"field":"email or password","issues":["Invalid"]}]}`, FixedTime),
		},
		{
			name:      "User not found",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, Email, Password),
			setupMock: func(s *MockAuth) {
				s.LoginFunc = func(ctx context.Context, email domain.Email, password domain.Password) (string, error) {
					return "", service.ErrUserNotFound // Enumeration and timing attacks are known, this is fine
				}
			},
			expectedCode: http.StatusUnauthorized,
			expectedBody: fmt.Sprintf(`{"code":"USER_NOT_FOUND","message":"User not found","timestamp":%q,"details":[]}`, FixedTime),
		},
		{
			name:      "Internal error",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, Email, Password),
			setupMock: func(s *MockAuth) {
				s.LoginFunc = func(ctx context.Context, email domain.Email, password domain.Password) (string, error) {
					return "", service.ErrInternal
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: fmt.Sprintf(`{"code":"INTERNAL_SERVER_ERROR","message":"Internal server error","timestamp":%q}`, FixedTime),
		},
		{
			name:         "Empty email",
			inputBody:    fmt.Sprintf(`{"email": %q, "password": %q}`, "", Password),
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"email","issues":["Invalid email"]}]}`, FixedTime),
		},
		{
			name:         "Invalid email format",
			inputBody:    fmt.Sprintf(`{"email": %q, "password": %q}`, "invalid-email", Password),
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"email","issues":["Invalid email"]}]}`, FixedTime),
		},
		{
			name:         "Empty email and password",
			inputBody:    fmt.Sprintf(`{"email": %q, "password": %q}`, "", ""),
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"email","issues":["Invalid email"]},{"field":"password","issues":["Password is too short"]}]}`, FixedTime),
		},
		{
			name:         "Empty password",
			inputBody:    fmt.Sprintf(`{"email": %q, "password": %q}`, Email, ""),
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"password","issues":["Password is too short"]}]}`, FixedTime),
		},
		{
			name:         "Invalid JSON",
			inputBody:    `{"email": "test@example.com"`, // missing password and closing brace
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"body","issues":["Invalid JSON body"]}]}`, FixedTime),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer([]byte(tt.inputBody)))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			s := &MockAuth{}

			if tt.setupMock != nil {
				tt.setupMock(s)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewAuth(logger, s, httpschema.MustNewErrorResponder(logger, MockTime))
			h.Login(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, rr.Code)
			}

			contentType := rr.Header().Get("Content-Type")
			mediaType, _, err := mime.ParseMediaType(contentType)
			if err != nil {
				t.Fatalf("Failed to parse MIME %q", contentType)
			}
			if mediaType != ExpectedMime {
				t.Errorf("Expected %q, got %q", ExpectedMime, mediaType)
			}

			testutil.AssertResponseBody(t, rr, tt.expectedBody)
		})
	}
}

func TestAuth_WhoAmI(t *testing.T) {
	t.Parallel()

	uid := testutil.ParseUserID("018e1000-0000-7000-8000-000000000000")
	ctx := context.WithValue(context.Background(), httpschema.ContextKeyUserID, uid)

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/whoami", nil)

	rr := httptest.NewRecorder()
	logger := testutil.NewTestLogger(t)
	h := handler.NewAuth(logger, &MockAuth{}, httpschema.MustNewErrorResponder(logger, MockTime))
	h.WhoAmI(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	expectedBody := fmt.Sprintf(`{"uid":%q}`, uid.String())

	testutil.AssertResponseBody(t, rr, expectedBody)
}
