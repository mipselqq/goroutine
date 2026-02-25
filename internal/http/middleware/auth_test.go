package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"goroutine/internal/http/handler"
	"goroutine/internal/http/middleware"
	"goroutine/internal/service"
	"goroutine/internal/testutil"
)

func TestAuth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		authorizationHeader string
		expectedStatus      int
		expectedUserID      int64
		setupMock           func(r *MockAuth)
	}{
		{
			name:                "Valid token",
			authorizationHeader: "Bearer valid.token.here",
			expectedStatus:      http.StatusTeapot,
			expectedUserID:      1,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 1, nil
				}
			},
		},
		{
			name:                "Invalid token",
			authorizationHeader: "Bearer invalid.token.here",
			expectedStatus:      http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 0, service.ErrInvalidToken
				}
			},
		},
		{
			name:                "Token expired",
			authorizationHeader: "Bearer expired.token.here",
			expectedStatus:      http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 0, service.ErrTokenExpired
				}
			},
		},
		{
			name:                "Invalid signing method",
			authorizationHeader: "Bearer invalid.signing.method.token.here",
			expectedStatus:      http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 0, service.ErrInvalidSigningMethod
				}
			},
		},
		{
			name:                "Missing header",
			authorizationHeader: "",
			expectedStatus:      http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 0, nil
				}
			},
		},
		{
			name:                "Missing token",
			authorizationHeader: "Bearer",
			expectedStatus:      http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 0, nil
				}
			},
		},
		{
			name:                "Empty token",
			authorizationHeader: "Bearer ",
			expectedStatus:      http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 0, nil
				}
			},
		},
		{
			name:                "Extra parts in header",
			authorizationHeader: "Bearer token extra-part",
			expectedStatus:      http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 0, nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &MockAuth{}
			tt.setupMock(s)

			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				id, ok := r.Context().Value(handler.UserIDKey).(int64)
				if !ok {
					t.Errorf("Expected user ID, got %v", id)
				}

				if id != tt.expectedUserId {
					t.Errorf("Expected user ID %d, got %d", tt.expectedUserId, id)
				}

				w.WriteHeader(http.StatusTeapot)
			})

			m := middleware.NewAuth(testutil.NewTestLogger(t), s)
			wrapped := m.Wrap(h)

			request := httptest.NewRequest("GET", "/", http.NoBody)
			request.Header.Set("Authorization", tt.authorizationHeader)
			response := httptest.NewRecorder()

			wrapped.ServeHTTP(response, request)

			if response.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, response.Code)
			}
		})
	}
}
