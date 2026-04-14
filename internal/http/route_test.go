package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	app "goroutine/internal/http"
	"goroutine/internal/http/handler"
	"goroutine/internal/http/httpschema"
	"goroutine/internal/http/middleware"
	"goroutine/internal/service"
	"goroutine/internal/testutil"
)

func TestNewRouter_Full(t *testing.T) {
	t.Parallel()

	logger := testutil.NewTestLogger(t)
	responder := httpschema.MustNewErrorResponder(logger, service.TimeNowRFC3339Millis)

	handlers := &handler.Handlers{
		Auth:   handler.NewAuth(logger, nil, responder),
		Health: handler.NewHealth(logger),
		Boards: handler.NewBoards(logger, nil, responder),
	}
	middlewares := &middleware.Middlewares{
		Metrics:   &spyMetricsMiddleware{},
		CORS:      &spyCorsMiddleware{},
		Auth:      &spyAuthMiddleware{},
		RequestID: &spyRequestIDMiddleware{},
	}

	router := app.NewRouter(handlers, middlewares)

	tests := []struct {
		name        string
		method      string
		path        string
		wantMetrics bool
		wantCors    bool
		wantReqID   bool
	}{ // TODO(refactor-1): use named fields, generate uuid
		{
			name:        "Register endpoint",
			method:      http.MethodPost,
			path:        "/v1/register",
			wantMetrics: true,
			wantCors:    true,
			wantReqID:   true,
		},
		{
			name:        "Login endpoint",
			method:      http.MethodPost,
			path:        "/v1/login",
			wantMetrics: true,
			wantCors:    true,
			wantReqID:   true,
		},
		{
			name:        "Health endpoint",
			method:      http.MethodGet,
			path:        "/v1/health",
			wantMetrics: true,
			wantCors:    true,
			wantReqID:   true,
		},
		{
			name:        "Boards list endpoint",
			method:      http.MethodGet,
			path:        "/v1/boards",
			wantMetrics: true,
			wantCors:    true,
			wantReqID:   true,
		},
		{
			name:        "Board by id endpoint",
			method:      http.MethodGet,
			path:        "/v1/boards/018e1000-0000-7000-8000-000000000001",
			wantMetrics: true,
			wantCors:    true,
			wantReqID:   true,
		},
		{
			name:        "UpdateByID board endpoint",
			method:      http.MethodPatch,
			path:        "/v1/boards/018e1000-0000-7000-8000-000000000001",
			wantMetrics: true,
			wantCors:    true,
			wantReqID:   true,
		},
		{
			name:        "Delete board endpoint",
			method:      http.MethodDelete,
			path:        "/v1/boards/018e1000-0000-7000-8000-000000000001",
			wantMetrics: true,
			wantCors:    true,
			wantReqID:   true,
		},
		{
			name:        "Swagger endpoint",
			method:      http.MethodGet,
			path:        "/v1/swagger/index.html",
			wantMetrics: false,
			wantCors:    true,
			wantReqID:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, http.NoBody)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			if rr.Code == http.StatusNotFound {
				t.Errorf("Path %q %q not registered in router (got 404)", tt.method, tt.path)
			}

			hasMetrics := rr.Header().Get("X-Metrics-Tracked") == "true"
			if hasMetrics != tt.wantMetrics {
				t.Errorf("Metrics middleware application mismatch for %q: got %v, want %v", tt.path, hasMetrics, tt.wantMetrics)
			}

			hasCors := rr.Header().Get("X-Cors-Tracked") == "true"
			if hasCors != tt.wantCors {
				t.Errorf("CORS middleware application mismatch for %q: got %v, want %v", tt.path, hasCors, tt.wantCors)
			}

			hasReqID := rr.Header().Get("X-RequestId-Tracked") == "true"
			if hasReqID != tt.wantReqID {
				t.Errorf("RequestID middleware application mismatch for %q: got %v, want %v", tt.path, hasReqID, tt.wantReqID)
			}
		})
	}

	t.Run("Non-existing endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/v1/non-existing", http.NoBody)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		testutil.AssertStatusCode(t, rr, http.StatusNotFound)
	})
}

func TestNewAdminRouter(t *testing.T) {
	t.Parallel()

	router := app.NewAdminRouter()

	req := httptest.NewRequest(http.MethodGet, "/metrics", http.NoBody)
	_, pattern := router.Handler(req)

	if pattern == "" {
		t.Errorf("Path GET /metrics not registered in admin router")
	}
}
