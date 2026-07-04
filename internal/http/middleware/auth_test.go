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
		name             string
		headerName       string
		headerValue      string
		wantStatus       int
		wantUserID       domain.UserID
		wantBody         any
		setupAuthService func(r *MockAuthService)
	}{
		{
			name:        "Valid token",
			headerName:  "Authorization",
			headerValue: "Bearer valid.token.here",
			wantStatus:  mockStatusCode,
			wantUserID:  userID,
			setupAuthService: func(r *MockAuthService) {
				r.VerifyTokenFunc = func(ctx context.Context, token domain.AuthToken) (domain.UserID, error) {
					return userID, nil
				}
			},
		},
		{
			name:        "Invalid token",
			headerName:  "Authorization",
			headerValue: "Bearer invalid.token.here",
			wantStatus:  http.StatusUnauthorized,
			setupAuthService: func(r *MockAuthService) {
				r.VerifyTokenFunc = func(ctx context.Context, token domain.AuthToken) (domain.UserID, error) {
					return domain.UserID{}, service.ErrInvalidToken
				}
			},
			wantBody: map[string]any{
				"code":      "INVALID_TOKEN",
				"message":   "Invalid token",
				"timestamp": testutil.FixedNowStr(),
				"details": []any{
					map[string]any{"field": "Authorization", "issues": []string{"Invalid token"}},
				},
			},
		},
		{
			name:        "Token expired",
			headerName:  "Authorization",
			headerValue: "Bearer expired.token.here",
			wantStatus:  http.StatusUnauthorized,
			setupAuthService: func(r *MockAuthService) {
				r.VerifyTokenFunc = func(ctx context.Context, token domain.AuthToken) (domain.UserID, error) {
					return domain.UserID{}, service.ErrTokenExpired
				}
			},
			wantBody: map[string]any{
				"code":      "INVALID_TOKEN",
				"message":   "Invalid token",
				"timestamp": testutil.FixedNowStr(),
				"details": []any{
					map[string]any{"field": "Authorization", "issues": []string{"Invalid token"}},
				},
			},
		},
		{
			name:        "Invalid signing method",
			headerName:  "Authorization",
			headerValue: "Bearer invalid.signing.method.token.here",
			wantStatus:  http.StatusUnauthorized,
			setupAuthService: func(r *MockAuthService) {
				r.VerifyTokenFunc = func(ctx context.Context, token domain.AuthToken) (domain.UserID, error) {
					return domain.UserID{}, service.ErrInvalidSigningMethod
				}
			},
			wantBody: map[string]any{
				"code":      "INVALID_TOKEN",
				"message":   "Invalid token",
				"timestamp": testutil.FixedNowStr(),
				"details": []any{
					map[string]any{"field": "Authorization", "issues": []string{"Invalid token"}},
				},
			},
		},
		{
			name:        "Empty header value",
			headerName:  "Authorization",
			headerValue: "",
			wantStatus:  http.StatusUnauthorized,
			setupAuthService: func(r *MockAuthService) {
				r.VerifyTokenFunc = func(ctx context.Context, token domain.AuthToken) (domain.UserID, error) {
					return domain.UserID{}, nil
				}
			},
			wantBody: map[string]any{
				"code":      "INVALID_AUTH_HEADER",
				"message":   "Invalid authorization header",
				"timestamp": testutil.FixedNowStr(),
				"details": []any{
					map[string]any{"field": "Authorization", "issues": []string{"Missing authorization header"}},
				},
			},
		},
		{
			name:        "Missing token",
			headerName:  "Authorization",
			headerValue: "Bearer",
			wantStatus:  http.StatusUnauthorized,
			setupAuthService: func(r *MockAuthService) {
				r.VerifyTokenFunc = func(ctx context.Context, token domain.AuthToken) (domain.UserID, error) {
					return domain.UserID{}, nil
				}
			},
			wantBody: map[string]any{
				"code":      "INVALID_AUTH_HEADER",
				"message":   "Invalid authorization header",
				"timestamp": testutil.FixedNowStr(),
				"details": []any{
					map[string]any{"field": "Authorization", "issues": []string{"Invalid authorization header"}},
				},
			},
		},
		{
			name:        "Empty token",
			headerName:  "Authorization",
			headerValue: "Bearer ",
			wantStatus:  http.StatusUnauthorized,
			setupAuthService: func(r *MockAuthService) {
				r.VerifyTokenFunc = func(ctx context.Context, token domain.AuthToken) (domain.UserID, error) {
					return domain.UserID{}, nil
				}
			},
			wantBody: map[string]any{
				"code":      "INVALID_AUTH_HEADER",
				"message":   "Invalid authorization header",
				"timestamp": testutil.FixedNowStr(),
				"details": []any{
					map[string]any{"field": "Authorization", "issues": []string{"Invalid authorization header"}},
				},
			},
		},
		{
			name:        "Extra parts in header",
			headerName:  "Authorization",
			headerValue: "Bearer token extra-part",
			wantStatus:  http.StatusUnauthorized,
			setupAuthService: func(r *MockAuthService) {
				r.VerifyTokenFunc = func(ctx context.Context, token domain.AuthToken) (domain.UserID, error) {
					return domain.UserID{}, nil
				}
			},
			wantBody: map[string]any{
				"code":      "INVALID_AUTH_HEADER",
				"message":   "Invalid authorization header",
				"timestamp": testutil.FixedNowStr(),
				"details": []any{
					map[string]any{"field": "Authorization", "issues": []string{"Invalid authorization header"}},
				},
			},
		},
		{
			name:        "Extra spaces in header",
			headerName:  "Authorization",
			headerValue: "   Bearer     valid.token.here   ",
			wantStatus:  mockStatusCode,
			wantUserID:  userID,
			setupAuthService: func(r *MockAuthService) {
				r.VerifyTokenFunc = func(ctx context.Context, token domain.AuthToken) (domain.UserID, error) {
					return userID, nil
				}
			},
		},
		{
			name:        "Random casing in header",
			headerName:  "AuThorIzAtIoN",
			headerValue: "bEaReR vAlId.tOkEn.hErE",
			wantStatus:  mockStatusCode,
			wantUserID:  userID,
			setupAuthService: func(r *MockAuthService) {
				r.VerifyTokenFunc = func(ctx context.Context, token domain.AuthToken) (domain.UserID, error) {
					return userID, nil
				}
			},
		},
		{
			name:        "Wrong prefix",
			headerName:  "Authorization",
			headerValue: "Basic some-token",
			wantStatus:  http.StatusUnauthorized,
			setupAuthService: func(r *MockAuthService) {
				r.VerifyTokenFunc = func(ctx context.Context, token domain.AuthToken) (domain.UserID, error) {
					return domain.UserID{}, nil
				}
			},
			wantBody: map[string]any{
				"code":      "INVALID_AUTH_HEADER",
				"message":   "Invalid authorization header",
				"timestamp": testutil.FixedNowStr(),
				"details": []any{
					map[string]any{"field": "Authorization", "issues": []string{"No Bearer prefix"}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := &MockAuthService{}
			tt.setupAuthService(s)

			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				id, ok := r.Context().Value(httpschema.ContextKeyUserID).(domain.UserID)
				if !ok {
					t.Errorf("got context user ID ok=%v, want true", ok)
				}

				if id != tt.wantUserID {
					t.Errorf("got user ID %v, want %v", id, tt.wantUserID)
				}

				w.WriteHeader(mockStatusCode)
			})

			logger := testutil.NewLogger(t)
			m := middleware.NewAuth(logger, s, httpschema.MustNewErrorResponder(logger, testutil.FixedNowStr))
			wrapped := m.Wrap(h)

			req, rr := testutil.NewJSONRequestAndRecorder(t, http.MethodGet, "/", "")

			req.Header.Set(tt.headerName, tt.headerValue)
			wrapped.ServeHTTP(rr, req)

			if rr.Code != mockStatusCode {
				testutil.AssertContentType(t, rr, "application/json")
			}
			testutil.AssertStatusCode(t, rr, tt.wantStatus)
			testutil.AssertResponseBody(t, rr, tt.wantBody)
		})
	}
}
