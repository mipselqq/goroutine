package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go-todo/internal/app"
	"go-todo/internal/handler"
	"go-todo/internal/testutil"
)

func TestNewRouter_Full(t *testing.T) {
	t.Parallel()

	logger := testutil.CreateTestLogger(t)
	authHandler := handler.NewAuth(logger, nil)
	healthHandler := handler.NewHealth(logger)

	router := app.NewRouter(authHandler, healthHandler)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"Register endpoint", http.MethodPost, "/register"},
		{"Login endpoint", http.MethodPost, "/login"},
		{"Health endpoint", http.MethodGet, "/health"},
		{"Swagger endpoint", http.MethodGet, "/swagger/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			_, pattern := router.Handler(req)

			if pattern == "" {
				t.Errorf("Path %s %s not registered in router", tt.method, tt.path)
			}
		})
	}

	t.Run("Non-existing endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/non-existing", nil)
		_, pattern := router.Handler(req)

		if pattern != "" {
			t.Errorf("Non-existing path registered in router")
		}
	})
}
