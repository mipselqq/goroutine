package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"goroutine/internal/http/httpschema"
	"goroutine/internal/http/middleware"
	"goroutine/internal/service"
	"goroutine/internal/testutil"
)

func TestAuth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		headerName     string
		headerValue    string
		expectedStatus int
		expectedUserID int64
		setupMock      func(r *MockAuth)
	}{
		{
			name:           "Valid token",
			headerName:     "Authorization",
			headerValue:    "Bearer valid.token.here",
			expectedStatus: http.StatusTeapot,
			expectedUserID: 1,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 1, nil
				}
			},
		},
		{
			name:           "Invalid token",
			headerName:     "Authorization",
			headerValue:    "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 0, service.ErrInvalidToken
				}
			},
		},
		{
			name:           "Token expired",
			headerName:     "Authorization",
			headerValue:    "Bearer expired.token.here",
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 0, service.ErrTokenExpired
				}
			},
		},
		{
			name:           "Invalid signing method",
			headerName:     "Authorization",
			headerValue:    "Bearer invalid.signing.method.token.here",
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 0, service.ErrInvalidSigningMethod
				}
			},
		},
		{
			name:           "Missing header",
			headerName:     "Authorization",
			headerValue:    "",
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 0, nil
				}
			},
		},
		{
			name:           "Missing token",
			headerName:     "Authorization",
			headerValue:    "Bearer",
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 0, nil
				}
			},
		},
		{
			name:           "Empty token",
			headerName:     "Authorization",
			headerValue:    "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 0, nil
				}
			},
		},
		{
			name:           "Extra parts in header",
			headerName:     "Authorization",
			headerValue:    "Bearer token extra-part",
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 0, nil
				}
			},
		},
		{
			name:           "Extra spaces in header",
			headerName:     "Authorization",
			headerValue:    "   Bearer     valid.token.here   ",
			expectedStatus: http.StatusTeapot,
			expectedUserID: 1,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 1, nil
				}
			},
		},
		{
			name:           "Random casing in header",
			headerName:     "AuThorIzAtIoN",
			headerValue:    "bEaReR vAlId.tOkEn.hErE",
			expectedStatus: http.StatusTeapot,
			expectedUserID: 1,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (int64, error) {
					return 1, nil
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
				id, ok := r.Context().Value(httpschema.ContextKeyUserID).(int64)
				if !ok {
					t.Errorf("Expected user ID, got %v", id)
				}

				if id != tt.expectedUserID {
					t.Errorf("Expected user ID %d, got %d", tt.expectedUserID, id)
				}

				w.WriteHeader(http.StatusTeapot)
			})

			m := middleware.NewAuth(testutil.NewTestLogger(t), s)
			wrapped := m.Wrap(h)

			request := httptest.NewRequest("GET", "/", http.NoBody)
			request.Header.Set(tt.headerName, tt.headerValue)
			response := httptest.NewRecorder()

			wrapped.ServeHTTP(response, request)

			if response.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, response.Code)
			}
		})
	}
}
