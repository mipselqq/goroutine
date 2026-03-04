package app

import (
	"log/slog"
	"net/http"

	"goroutine/internal/config"
	"goroutine/internal/domain"
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

func New(logger *slog.Logger, pool *pgxpool.Pool, cfg *config.AppConfig) *App {
	userRepo := repository.NewPgUser(pool)
	authService := service.NewAuth(userRepo, service.JWTOptions{
		JWTSecret:     cfg.JWTSecret,
		Exp:           cfg.JWTExp,
		SigningMethod: jwt.SigningMethodHS256,
	})

	responder := httpschema.MustNewErrorResponder(logger, service.TimeRFC3339Milli)

	authHandler := handler.NewAuth(logger, authService, responder)
	healthHandler := handler.NewHealth(logger)

	metricsMiddleware := middleware.NewMetrics(prometheus.DefaultRegisterer)
	corsMiddleware := middleware.NewCORS(logger, cfg.AllowedOrigins)
	authMiddleware := middleware.NewAuth(logger, authService, responder)
	reqIDMiddleware := middleware.NewRequestID(logger, func() string {
		return domain.NewUserID().String()
	})

	handlers := &handler.Handlers{Auth: authHandler, Health: healthHandler}
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
