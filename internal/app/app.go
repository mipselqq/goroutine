package app

import (
	"log/slog"
	"net/http"

	"goroutine/internal/config"
	"goroutine/internal/handler"
	"goroutine/internal/repository"
	"goroutine/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(logger *slog.Logger, pool *pgxpool.Pool, cfg *config.AppConfig) http.Handler {
	userRepo := repository.NewPgUser(pool)
	authService := service.NewAuth(userRepo, service.JWTOptions{
		JWTSecret: cfg.JWTSecret,
		Exp:       cfg.JWTExp,
	})

	authHandler := handler.NewAuth(logger, authService)
	healthHandler := handler.NewHealth(logger)

	return NewRouter(authHandler, healthHandler)
}
