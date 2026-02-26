package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	app "goroutine/internal/http"
	"goroutine/internal/http/handler"
	"goroutine/internal/http/middleware"
	"goroutine/internal/testutil"
)

func TestNewRouter_Full(t *testing.T) {
	t.Parallel()

	logger := testutil.NewTestLogger(t)

	handlers := &handler.Handlers{
		Auth:   handler.NewAuth(logger, nil),
		Health: handler.NewHealth(logger),
	}
	middlewares := &middleware.Middlewares{
		Metrics: &spyMetricsMiddleware{},
		CORS:    &spyCorsMiddleware{},
		Auth:    &spyAuthMiddleware{},
	}

	router := app.NewRouter(handlers, middlewares)

	tests := []struct {
		name        string
		method      string
		path        string
		wantMetrics bool
		wantCors    bool
	}{
		{"Register endpoint", http.MethodPost, "/register", true, true},
		{"Login endpoint", http.MethodPost, "/login", true, true},
		{"Health endpoint", http.MethodGet, "/health", true, true},
		{"Swagger endpoint", http.MethodGet, "/swagger/index.html", false, true},
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
		})
	}

	t.Run("Non-existing endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/non-existing", http.NoBody)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("Expected 404 for non-existing endpoint, got %d", rr.Code)
		}
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
