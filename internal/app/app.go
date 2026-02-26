package app

import (
	"log/slog"
	"net/http"

	"goroutine/internal/config"
	httpapp "goroutine/internal/http"
	"goroutine/internal/http/handler"
	"goroutine/internal/http/middleware"
	"goroutine/internal/repository"
	"goroutine/internal/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

type App struct {
	Router      http.Handler
	AdminRouter http.Handler
}

func New(logger *slog.Logger, pool *pgxpool.Pool, cfg *config.AppConfig) *App {
	userRepo := repository.NewPgUser(pool)
	authService := service.NewAuth(userRepo, service.JWTOptions{
		JWTSecret:     cfg.JWTSecret,
		Exp:           cfg.JWTExp,
		SigningMethod: jwt.SigningMethodHS256,
	})

	authHandler := handler.NewAuth(logger, authService)
	healthHandler := handler.NewHealth(logger)

	metricsMiddleware := middleware.NewMetrics(prometheus.DefaultRegisterer)
	corsMiddleware := middleware.NewCORS(logger, cfg.AllowedOrigins)
	authMiddleware := middleware.NewAuth(logger, authService)

	return &App{
		Router:      httpapp.NewRouter(metricsMiddleware, corsMiddleware, authMiddleware, authHandler, healthHandler),
		AdminRouter: httpapp.NewAdminRouter(),
	}
}
