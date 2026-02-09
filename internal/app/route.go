package app

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"goroutine/internal/handler"
)

func NewRouter(authHandler *handler.Auth, healthHandler *handler.Health) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("GET /health", healthHandler.Health)
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)
	mux.Handle("GET /metrics", promhttp.Handler())

	return mux
}
