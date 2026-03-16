package app

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"goroutine/internal/config"
	httpapp "goroutine/internal/http"
	"goroutine/internal/http/handler"
	"goroutine/internal/http/httpschema"
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

func New(logger *slog.Logger, pool *pgxpool.Pool, cfg *config.AppConfig, reg prometheus.Registerer) *App {
	userRepo := repository.NewPgUser(pool)
	authService := service.NewAuth(userRepo, service.JWTOptions{
		JWTSecret:     cfg.JWTSecret,
		Exp:           cfg.JWTExp,
		SigningMethod: jwt.SigningMethodHS256,
	})

	responder := httpschema.MustNewErrorResponder(logger, service.TimeRFC3339Milli)

	authHandler := handler.NewAuth(logger, authService, responder)
	healthHandler := handler.NewHealth(logger)

	boardRepo := repository.NewPgBoard(pool)
	boardsService := service.NewBoard(boardRepo)
	boardsHandler := handler.NewBoards(logger, boardsService, responder)

	metricsMiddleware := middleware.NewMetrics(reg)
	corsMiddleware := middleware.NewCORS(logger, cfg.AllowedOrigins)
	authMiddleware := middleware.NewAuth(logger, authService, responder)
	reqIDMiddleware := middleware.MustNewRequestID(logger, func() string {
		return fmt.Sprintf("req-%s", uuid.Must(uuid.NewV7()))
	})

	handlers := &handler.Handlers{Auth: authHandler, Health: healthHandler, Boards: boardsHandler}
	middlewares := &middleware.Middlewares{
		Metrics:   metricsMiddleware,
		CORS:      corsMiddleware,
		Auth:      authMiddleware,
		RequestID: reqIDMiddleware,
	}

	return &App{
		Router:      httpapp.NewRouter(handlers, middlewares),
		AdminRouter: httpapp.NewAdminRouter(),
	}
}
