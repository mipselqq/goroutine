package app

import (
	"net/http"

	"goroutine/internal/http/handler"
	"goroutine/internal/http/middleware"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func NewRouter(h *handler.Handlers, m *middleware.Middlewares) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /register", m.Metrics.Wrap(http.HandlerFunc(h.Auth.Register)))
	mux.Handle("POST /login", m.Metrics.Wrap(http.HandlerFunc(h.Auth.Login)))
	mux.Handle("GET /health", m.Metrics.Wrap(http.HandlerFunc(h.Health.Health)))
	mux.Handle("GET /whoami", m.Metrics.Wrap(m.Auth.Wrap(http.HandlerFunc(h.Auth.WhoAmI))))
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	// Apply CORS to all routes
	return m.CORS.Wrap(mux)
}

func NewAdminRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())

	return mux
}
