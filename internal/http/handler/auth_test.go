package handler_test

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goroutine/internal/domain"
	"goroutine/internal/http/handler"
	"goroutine/internal/http/httpschema"
	"goroutine/internal/service"
	"goroutine/internal/testutil"
)

type TestCase struct {
	name         string
	inputBody    string
	setupMock    func(s *MockAuth)
	expectedCode int
	expectedBody string
}

const (
	email        string = "test@example.com"
	password     string = "qwerty"
	expectedMime string = "application/json"
)

func TestAuth_Register(t *testing.T) {
	t.Parallel()

	tests := []TestCase{
		{
			name:      "Success",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, email, password),
			setupMock: func(s *MockAuth) {
				s.RegisterFunc = func(ctx context.Context, e domain.Email, p domain.Password) error {
					if e.String() != email {
						t.Errorf("expected email %s, got %s", email, e.String())
					}
					if p.String() != password {
						t.Errorf("expected password %s, got %s", password, p.String())
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
			expectedBody: `{"error":"internal error happened"}`,
		},
		{
			name:         "Empty email",
			inputBody:    fmt.Sprintf(`{"email": %q, "password": %q}`, "", password),
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"invalid email"}`,
		},
		{
			name:         "Invalid email format",
			inputBody:    fmt.Sprintf(`{"email": %q, "password": %q}`, "invalid-email", password),
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"invalid email"}`,
		},
		{
			name:         "Empty password",
			inputBody:    fmt.Sprintf(`{"email": %q, "password": %q}`, email, ""),
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"invalid password"}`,
		},
		{
			name:         "Password too short",
			inputBody:    fmt.Sprintf(`{"email": %q, "password": %q}`, email, "123"),
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"password must be at least 6 characters"}`,
		},
		{
			name:         "Invalid JSON",
			inputBody:    `{"email": "test@example.com", "password": "password"`, // missing closing brace
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"invalid json body"}`,
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
			expectedBody: `{"error":"user already exists"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte(tt.inputBody)))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			s := MockAuth{}

			if tt.setupMock != nil {
				tt.setupMock(&s)
			}

			h := handler.NewAuth(testutil.NewTestLogger(t), &s)
			h.Register(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, rr.Code)
			}

			contentType := rr.Header().Get("Content-Type")
			mediaType, _, err := mime.ParseMediaType(contentType)
			if err != nil {
				t.Fatalf("Failed to parse MIME %s", contentType)
			}
			if mediaType != expectedMime {
				t.Errorf("Expected %s, got %s", expectedMime, mediaType)
			}

			if tt.expectedBody != "" {
				actualBody := bytes.TrimSpace(rr.Body.Bytes())
				if string(actualBody) != tt.expectedBody {
					t.Errorf("expected body %s, got %s", tt.expectedBody, string(actualBody))
				}
			}
		})
	}
}

func TestAuth_Login(t *testing.T) {
	t.Parallel()

	tests := []TestCase{
		{
			name:      "Success",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, email, password),
			setupMock: func(s *MockAuth) {
				s.LoginFunc = func(ctx context.Context, e domain.Email, p domain.Password) (string, error) {
					if e.String() != email {
						t.Errorf("expected email %s, got %s", email, e.String())
					}
					if p.String() != password {
						t.Errorf("expected password %s, got %s", password, p.String())
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
			expectedBody: `{"error":"invalid email or password"}`,
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
			expectedBody: `{"error":"user not found"}`,
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
			expectedBody: `{"error":"internal error happened"}`,
		},
		{
			name:         "Empty email",
			inputBody:    fmt.Sprintf(`{"email": %q, "password": %q}`, "", password),
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"invalid email"}`,
		},
		{
			name:         "Invalid email format",
			inputBody:    fmt.Sprintf(`{"email": %q, "password": %q}`, "invalid-email", password),
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"invalid email"}`,
		},
		{
			name:         "Empty password",
			inputBody:    fmt.Sprintf(`{"email": %q, "password": %q}`, email, ""),
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"invalid password"}`,
		},
		{
			name:         "Invalid JSON",
			inputBody:    `{"email": "test@example.com"`, // missing password and closing brace
			expectedCode: http.StatusBadRequest,
			expectedBody: `{"error":"invalid json body"}`,
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

			h := handler.NewAuth(testutil.NewTestLogger(t), s)
			h.Login(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, rr.Code)
			}

			contentType := rr.Header().Get("Content-Type")
			mediaType, _, err := mime.ParseMediaType(contentType)
			if err != nil {
				t.Fatalf("Failed to parse MIME %s", contentType)
			}
			if mediaType != expectedMime {
				t.Errorf("Expected %s, got %s", expectedMime, mediaType)
			}

			if tt.expectedBody != "" {
				actualBody := bytes.TrimSpace(rr.Body.Bytes())
				if string(actualBody) != tt.expectedBody {
					t.Errorf("expected body %s, got %s", tt.expectedBody, string(actualBody))
				}
			}
		})
	}
}

func TestAuth_WhoAmI(t *testing.T) {
	t.Parallel()

	uid := domain.MustParseUserID("018e1000-0000-7000-8000-000000000000")
	ctx := context.WithValue(context.Background(), httpschema.ContextKeyUserID, uid)

	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/whoami", nil)

	rr := httptest.NewRecorder()
	h := handler.NewAuth(testutil.NewTestLogger(t), &MockAuth{})
	h.WhoAmI(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	expectedBody := fmt.Sprintf(`{"uid":%q}`, uid.String())
	if strings.TrimSpace(rr.Body.String()) != expectedBody {
		t.Errorf("expected body %q, got %q", expectedBody, rr.Body.String())
	}
}
