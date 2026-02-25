package app

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"goroutine/internal/http/handler"
)

type Middleware interface {
	Wrap(next http.Handler) http.HandlerFunc
}

func NewRouter(metricsMiddleware, corsMiddleware, authMiddleware Middleware, authHandler *handler.Auth, healthHandler *handler.Health) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("POST /register", metricsMiddleware.Wrap(http.HandlerFunc(authHandler.Register)))
	mux.Handle("POST /login", metricsMiddleware.Wrap(http.HandlerFunc(authHandler.Login)))
	mux.Handle("GET /health", metricsMiddleware.Wrap(http.HandlerFunc(healthHandler.Health)))
	mux.Handle("GET /whoami", metricsMiddleware.Wrap(authMiddleware.Wrap(http.HandlerFunc(authHandler.WhoAmI))))
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	return corsMiddleware.Wrap(mux)
}

func NewAdminRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())

	return mux
}
