package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"goroutine/internal/app"
	"goroutine/internal/handler"
	"goroutine/internal/testutil"
)

type spyMetricsMiddleware struct{}

func (s *spyMetricsMiddleware) Wrap(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Metrics-Tracked", "true")
		next.ServeHTTP(w, r)
	}
}

func TestNewRouter_Full(t *testing.T) {
	t.Parallel()

	logger := testutil.CreateTestLogger(t)
	authHandler := handler.NewAuth(logger, nil)
	healthHandler := handler.NewHealth(logger)
	metricsMiddleware := &spyMetricsMiddleware{}

	router := app.NewRouter(metricsMiddleware, authHandler, healthHandler)

	tests := []struct {
		name        string
		method      string
		path        string
		wantMetrics bool
	}{
		{"Register endpoint", http.MethodPost, "/register", true},
		{"Login endpoint", http.MethodPost, "/login", true},
		{"Health endpoint", http.MethodGet, "/health", true},
		{"Swagger endpoint", http.MethodGet, "/swagger/index.html", false},
		{"Metrics endpoint", http.MethodGet, "/metrics", false},
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
