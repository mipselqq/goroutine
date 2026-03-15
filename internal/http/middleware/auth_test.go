package middleware_test

import (
	"context"
	"net/http"
	"testing"

	"goroutine/internal/domain"
	"goroutine/internal/http/httpschema"
	"goroutine/internal/http/middleware"
	"goroutine/internal/service"
	"goroutine/internal/testutil"
)

func TestAuth(t *testing.T) {
	t.Parallel()

	userID := testutil.ValidUserID()
	mockStatusCode := http.StatusTeapot

	tests := []struct {
		name           string
		headerName     string
		headerValue    string
		expectedStatus int
		expectedUserID domain.UserID
		expectedBody   any
		setupMock      func(r *MockAuth)
	}{
		{
			name:           "Valid token",
			headerName:     "Authorization",
			headerValue:    "Bearer valid.token.here",
			expectedStatus: mockStatusCode,
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
			expectedBody: map[string]any{
				"code":      "INVALID_TOKEN",
				"message":   "Invalid token",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "Authorization", "issues": []string{"Invalid token"}},
				},
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
			expectedBody: map[string]any{
				"code":      "INVALID_TOKEN",
				"message":   "Invalid token",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "Authorization", "issues": []string{"Invalid token"}},
				},
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
			expectedBody: map[string]any{
				"code":      "INVALID_TOKEN",
				"message":   "Invalid token",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "Authorization", "issues": []string{"Invalid token"}},
				},
			},
		},
		{
			name:           "Empty header value",
			headerName:     "Authorization",
			headerValue:    "",
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (domain.UserID, error) {
					return domain.UserID{}, nil
				}
			},
			expectedBody: map[string]any{
				"code":      "INVALID_AUTH_HEADER",
				"message":   "Invalid authorization header",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "Authorization", "issues": []string{"Missing authorization header"}},
				},
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
			expectedBody: map[string]any{
				"code":      "INVALID_AUTH_HEADER",
				"message":   "Invalid authorization header",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "Authorization", "issues": []string{"Invalid authorization header"}},
				},
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
			expectedBody: map[string]any{
				"code":      "INVALID_AUTH_HEADER",
				"message":   "Invalid authorization header",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "Authorization", "issues": []string{"Invalid authorization header"}},
				},
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
			expectedBody: map[string]any{
				"code":      "INVALID_AUTH_HEADER",
				"message":   "Invalid authorization header",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "Authorization", "issues": []string{"Invalid authorization header"}},
				},
			},
		},
		{
			name:           "Extra spaces in header",
			headerName:     "Authorization",
			headerValue:    "   Bearer     valid.token.here   ",
			expectedStatus: mockStatusCode,
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
			expectedStatus: mockStatusCode,
			expectedUserID: userID,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (domain.UserID, error) {
					return userID, nil
				}
			},
		},
		{
			name:           "Wrong prefix",
			headerName:     "Authorization",
			headerValue:    "Basic some-token",
			expectedStatus: http.StatusUnauthorized,
			setupMock: func(r *MockAuth) {
				r.VerifyTokenFunc = func(ctx context.Context, token string) (domain.UserID, error) {
					return domain.UserID{}, nil
				}
			},
			expectedBody: map[string]any{
				"code":      "INVALID_AUTH_HEADER",
				"message":   "Invalid authorization header",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "Authorization", "issues": []string{"No Bearer prefix"}},
				},
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

				w.WriteHeader(mockStatusCode)
			})

			logger := testutil.NewTestLogger(t)
			m := middleware.NewAuth(logger, s, httpschema.MustNewErrorResponder(logger, testutil.FixedTime))
			wrapped := m.Wrap(h)

			req, rr := testutil.NewJSONRequestAndRecorder(t, http.MethodGet, "/", "")

			req.Header.Set(tt.headerName, tt.headerValue)
			wrapped.ServeHTTP(rr, req)

			if rr.Code != mockStatusCode {
				testutil.AssertContentType(t, rr, "application/json")
			}
			testutil.AssertStatusCode(t, rr, tt.expectedStatus)
			testutil.AssertResponseBody(t, rr, tt.expectedBody)
		})
	}
}
