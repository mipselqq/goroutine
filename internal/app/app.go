// Package app wires the application layers together, including dependency injection.
package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/google/uuid"

	"goroutine/internal/config"
	"goroutine/internal/domain"
	telegramDrv "goroutine/internal/driver/telegram"
	httpapp "goroutine/internal/http"
	"goroutine/internal/http/handler"
	"goroutine/internal/http/httpschema"
	"goroutine/internal/http/middleware"
	"goroutine/internal/repository"
	"goroutine/internal/service"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

type App struct {
	Router      http.Handler
	AdminRouter http.Handler
}

func New(
	logger *slog.Logger,
	pgPool *pgxpool.Pool,
	redisClient *redis.Client,
	cfg *config.AppConfig,
	telegramCfg *config.TelegramConfig,
	reg prometheus.Registerer,
) *App {
	userRepo := repository.NewPgUser(pgPool)
	telegramTokenRepo := repository.NewRedisTelegramToken(redisClient, telegramCfg.LinkTokenTTL)
	boardsRepo := repository.NewPgBoard(pgPool)
	columnsRepo := repository.NewPgColumn(pgPool)
	tasksRepo := repository.NewPgTask(pgPool)

	authService := service.NewAuth(userRepo, service.JWTOptions{
		JWTSecret:     cfg.JWTSecret,
		Exp:           cfg.JWTExp,
		SigningMethod: jwt.SigningMethodHS256,
	})
	telegramAPIClient := telegramDrv.NewAPIClient(telegramCfg.BaseURL, telegramCfg.Token)
	userService := service.NewUser(userRepo, telegramTokenRepo, func() domain.TelegramLinkToken {
		tok, err := domain.NewTelegramLinkToken(uuid.Must(uuid.NewV7()).String())
		if err != nil {
			logger.Error(fmt.Sprintf("BUG: NewTelegramLinkToken() rejected UUIDv7: %v", err))
			os.Exit(1)
		}
		return tok
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
	userHandler := handler.NewUser(logger, userService, errorResponder)
	telegramWebhook := telegramDrv.NewWebhookHandler(logger, userService, telegramAPIClient)

	metricsMiddleware := middleware.NewMetrics(reg)
	corsMiddleware := middleware.NewCORS(logger, cfg.AllowedOrigins)
	authMiddleware := middleware.NewAuth(logger, authService, errorResponder)
	reqIDMiddleware := middleware.MustNewRequestID(logger, func() string {
		return fmt.Sprintf("req-%s", uuid.Must(uuid.NewV7()))
	})

	handlers := &handler.Handlers{
		Auth:    authHandler,
		Health:  healthHandler,
		Boards:  boardsHandler,
		Columns: columnsHandler,
		Tasks:   tasksHandler,
		User:    userHandler,
	}
	middlewares := &middleware.Middlewares{
		Metrics:   metricsMiddleware,
		CORS:      corsMiddleware,
		Auth:      authMiddleware,
		RequestID: reqIDMiddleware,
	}

	return &App{
		Router:      httpapp.NewRouter(handlers, middlewares, telegramWebhook),
		AdminRouter: httpapp.NewAdminRouter(),
	}
}
