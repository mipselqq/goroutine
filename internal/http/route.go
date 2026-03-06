package http

import (
	"net/http"

	"goroutine/internal/http/handler"
	"goroutine/internal/http/middleware"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

func NewRouter(h *handler.Handlers, m *middleware.Middlewares) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /v1/register", m.Metrics.Wrap(http.HandlerFunc(h.Auth.Register)))
	mux.Handle("POST /v1/login", m.Metrics.Wrap(http.HandlerFunc(h.Auth.Login)))
	mux.Handle("GET /v1/health", m.Metrics.Wrap(http.HandlerFunc(h.Health.Health)))
	mux.Handle("GET /v1/whoami", m.Metrics.Wrap(m.Auth.Wrap(http.HandlerFunc(h.Auth.WhoAmI))))
	mux.Handle("GET /v1/swagger/", httpSwagger.WrapHandler)

	return m.RequestID.Wrap(m.CORS.Wrap(mux))
}

func NewAdminRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())

	return mux
}
