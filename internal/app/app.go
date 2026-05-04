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
	boardsRepo := repository.NewPgBoard(pool)
	columnsRepo := repository.NewPgColumn(pool)
	tasksRepo := repository.NewPgTask(pool)

	authService := service.NewAuth(userRepo, service.JWTOptions{
		JWTSecret:     cfg.JWTSecret,
		Exp:           cfg.JWTExp,
		SigningMethod: jwt.SigningMethodHS256,
	})
	boardsService := service.NewBoard(boardsRepo, columnsRepo, tasksRepo)
	columnsService := service.NewColumn(columnsRepo, boardsRepo)
	tasksService := service.NewTask(tasksRepo, boardsRepo, columnsRepo)

	errorResponder := httpschema.MustNewErrorResponder(logger, service.TimeNowRFC3339Millis)
	authHandler := handler.NewAuth(logger, authService, errorResponder)
	healthHandler := handler.NewHealth(logger)
	boardsHandler := handler.NewBoards(logger, boardsService, errorResponder)
	columnsHandler := handler.NewColumns(logger, columnsService, errorResponder)
	tasksHandler := handler.NewTasks(logger, tasksService, errorResponder)

	metricsMiddleware := middleware.NewMetrics(reg)
	corsMiddleware := middleware.NewCORS(logger, cfg.AllowedOrigins)
	authMiddleware := middleware.NewAuth(logger, authService, errorResponder)
	reqIDMiddleware := middleware.MustNewRequestID(logger, func() string {
		return fmt.Sprintf("req-%s", uuid.Must(uuid.NewV7()))
	})

	handlers := &handler.Handlers{Auth: authHandler, Health: healthHandler, Boards: boardsHandler, Columns: columnsHandler, Tasks: tasksHandler}
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
