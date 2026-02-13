package app

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"goroutine/internal/handler"
)

type Middleware interface {
	Wrap(next http.Handler) http.HandlerFunc
}

func NewRouter(metricsMiddleware Middleware, authHandler *handler.Auth, healthHandler *handler.Health) *http.ServeMux {
	mux := http.NewServeMux()

	mux.Handle("POST /register", metricsMiddleware.Wrap(http.HandlerFunc(authHandler.Register)))
	mux.HandleFunc("POST /login", metricsMiddleware.Wrap(http.HandlerFunc(authHandler.Login)))
	mux.HandleFunc("GET /health", metricsMiddleware.Wrap(http.HandlerFunc(healthHandler.Health)))
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)
	mux.Handle("GET /metrics", promhttp.Handler())

	return mux
}
