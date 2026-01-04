package handler_test

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-todo/internal/domain"
	"go-todo/internal/handler"
	"go-todo/internal/service"
	"go-todo/internal/testutils"
)

type MockAuthService struct {
	RegisterFunc func(ctx context.Context, email domain.Email, password domain.Password) error
}

func (m *MockAuthService) Register(ctx context.Context, email domain.Email, password domain.Password) error {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, email, password)
	}
	return nil
}

func TestAuth_Register(t *testing.T) {
	t.Parallel()

	email := "test@example.com"
	password := "qwerty"
	expectedMime := "application/json"

	tests := []struct {
		name         string
		inputBody    string
		setupMock    func(s *MockAuthService)
		expectedCode int
	}{
		{
			name:      "Success",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, email, password),
			setupMock: func(s *MockAuthService) {
				s.RegisterFunc = func(ctx context.Context, email domain.Email, password domain.Password) error {
					return nil
				}
			},
			expectedCode: http.StatusOK,
		},
		{
			name:      "Internal error happened for valid body",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, email, password),
			setupMock: func(s *MockAuthService) {
				s.RegisterFunc = func(ctx context.Context, email domain.Email, password domain.Password) error {
					return service.ErrInternal
				}
			},
			expectedCode: http.StatusInternalServerError,
		},
		{
			name:      "Empty email",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, "", password),
			setupMock: func(s *MockAuthService) {
				s.RegisterFunc = func(ctx context.Context, email domain.Email, password domain.Password) error {
					t.Error("Service should not be called")
					return nil
				}
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:      "Empty password",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, email, ""),
			setupMock: func(s *MockAuthService) {
				s.RegisterFunc = func(ctx context.Context, email domain.Email, password domain.Password) error {
					t.Error("Service should not be called")
					return nil
				}
			},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:         "Invalid JSON",
			inputBody:    fmt.Sprintf(`{email: %q, "password": %q}`, email, password),
			setupMock:    func(s *MockAuthService) {},
			expectedCode: http.StatusBadRequest,
		},
		{
			name:      "User already exists",
			inputBody: fmt.Sprintf(`{"email": %q, "password": %q}`, email, password),
			setupMock: func(s *MockAuthService) {
				s.RegisterFunc = func(ctx context.Context, email domain.Email, password domain.Password) error {
					return service.ErrUserAlreadyExists
				}
			},
			expectedCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte(tt.inputBody)))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			s := &MockAuthService{}

			if tt.setupMock != nil {
				tt.setupMock(s)
			}

			h := handler.NewAuth(testutils.CreateTestLogger(t), s)
			h.Register(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, rr.Code)
			}

			// TODO: send nosniff header in middleware
			contentType := rr.Header().Get("Content-Type")
			mediaType, _, err := mime.ParseMediaType(contentType)
			if err != nil {
				t.Fatalf("Failed to parse MIME %s", contentType)
			}
			if mediaType != expectedMime {
				t.Errorf("Expected %s, got %s", expectedMime, mediaType)
			}
		})
	}
}
