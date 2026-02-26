package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"goroutine/internal/domain"
	"goroutine/internal/http/httpschema"
	"goroutine/internal/http/middleware"
	"goroutine/internal/service"
	"goroutine/internal/testutil"
)

func TestAuth(t *testing.T) {
	t.Parallel()

	userID := testutil.ParseUserID("018e1000-0000-7000-8000-000000000000")

	tests := []struct {
		name           string
		headerName     string
		headerValue    string
		expectedStatus int
		expectedUserID domain.UserID
		setupMock      func(r *MockAuth)
	}{
		{
			name:           "Valid token",
			headerName:     "Authorization",
			headerValue:    "Bearer valid.token.here",
			expectedStatus: http.StatusTeapot,
			expectedUserID: userID,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (domain.UserID, error) {
					return userID, nil
				}
			},
		},
		{
			name:           "Invalid token",
			headerName:     "Authorization",
			headerValue:    "Bearer invalid.token.here",
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (domain.UserID, error) {
					return domain.UserID{}, service.ErrInvalidToken
				}
			},
		},
		{
			name:           "Token expired",
			headerName:     "Authorization",
			headerValue:    "Bearer expired.token.here",
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (domain.UserID, error) {
					return domain.UserID{}, service.ErrTokenExpired
				}
			},
		},
		{
			name:           "Invalid signing method",
			headerName:     "Authorization",
			headerValue:    "Bearer invalid.signing.method.token.here",
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (domain.UserID, error) {
					return domain.UserID{}, service.ErrInvalidSigningMethod
				}
			},
		},
		{
			name:           "Missing header",
			headerName:     "Authorization",
			headerValue:    "",
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (domain.UserID, error) {
					return domain.UserID{}, nil
				}
			},
		},
		{
			name:           "Missing token",
			headerName:     "Authorization",
			headerValue:    "Bearer",
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (domain.UserID, error) {
					return domain.UserID{}, nil
				}
			},
		},
		{
			name:           "Empty token",
			headerName:     "Authorization",
			headerValue:    "Bearer ",
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (domain.UserID, error) {
					return domain.UserID{}, nil
				}
			},
		},
		{
			name:           "Extra parts in header",
			headerName:     "Authorization",
			headerValue:    "Bearer token extra-part",
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (domain.UserID, error) {
					return domain.UserID{}, nil
				}
			},
		},
		{
			name:           "Extra spaces in header",
			headerName:     "Authorization",
			headerValue:    "   Bearer     valid.token.here   ",
			expectedStatus: http.StatusTeapot,
			expectedUserID: userID,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (domain.UserID, error) {
					return userID, nil
				}
			},
		},
		{
			name:           "Random casing in header",
			headerName:     "AuThorIzAtIoN",
			headerValue:    "bEaReR vAlId.tOkEn.hErE",
			expectedStatus: http.StatusTeapot,
			expectedUserID: userID,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (domain.UserID, error) {
					return userID, nil
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
				id, ok := r.Context().Value(httpschema.ContextKeyUserID).(domain.UserID)
				if !ok {
					t.Errorf("Expected user ID, got %v", id)
				}

				if id != tt.expectedUserID {
					t.Errorf("Expected user ID %v, got %v", tt.expectedUserID, id)
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
