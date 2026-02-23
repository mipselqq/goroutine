package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"goroutine/internal/app"
	"goroutine/internal/handler"
	"goroutine/internal/testutil"
)

func TestNewRouter_Full(t *testing.T) {
	t.Parallel()

	logger := testutil.NewTestLogger(t)
	authHandler := handler.NewAuth(logger, nil)
	healthHandler := handler.NewHealth(logger)
	metricsMiddleware := &spyMetricsMiddleware{}
	corsMiddleware := &spyCorsMiddleware{}

	router := app.NewRouter(metricsMiddleware, corsMiddleware, authHandler, healthHandler)

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
		{"Swagger endpoint", http.MethodGet, "/swagger/index.html", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, http.NoBody)
			rr := httptest.NewRecorder()

			_, pattern := router.Handler(req)
			if pattern == "" {
				t.Errorf("Path %s %s not registered in router", tt.method, tt.path)
			}

			router.ServeHTTP(rr, req)
			hasMetrics := rr.Header().Get("X-Metrics-Tracked") == "true"
			if hasMetrics != tt.wantMetrics {
				t.Errorf("Metrics middleware application mismatch for %s: got %v, want %v", tt.path, hasMetrics, tt.wantMetrics)
			}
			hasCors := rr.Header().Get("X-Cors-Tracked") == "true"
			if hasCors != tt.wantCors {
				t.Errorf("CORS middleware not applied to %s ", tt.path)
			}
			if hasCors != tt.wantCors {
				t.Errorf("CORS middleware application mismatch for %s: got %v, want %v", tt.path, hasCors, tt.wantCors)
			}
		})
	}

	t.Run("Non-existing endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/non-existing", http.NoBody)
		_, pattern := router.Handler(req)

		if pattern != "" {
			t.Errorf("Non-existing path registered in router")
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
