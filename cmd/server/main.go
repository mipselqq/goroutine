package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"go-todo/internal/config"
	"go-todo/internal/handler"
	"go-todo/internal/logging"
	"go-todo/internal/repository"
	"go-todo/internal/service"
)

type authServiceAdapter struct {
	service *service.Auth
}

func (a *authServiceAdapter) Register(ctx context.Context, email, password string) (string, error) {
	if err := a.service.Register(ctx, email, password); err != nil {
		return "", err
	}

	return "simulated_token", nil
}

func main() {
	appCfg := config.NewAppConfigFromEnv()
	logger := logging.NewLogger(appCfg.Env, appCfg.LogLevel)

	cfg := config.NewPGConfigFromEnv()

	logger.Info("Database config", slog.Any("config", cfg))
	logger.Info("App config", slog.Any("config", appCfg))

	poolConfig, err := cfg.ParsePGXpoolConfig()
	if err != nil {
		logger.Error("Failed to parse database config", slog.String("err", err.Error()))
		os.Exit(1)
	}

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Error("Failed to connect to database", slog.String("err", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Error("Failed to ping database", slog.String("err", err.Error()))
		os.Exit(1)
	}

	userRepo := repository.NewPgUser(pool)
	authService := service.NewAuth(userRepo)
	authAdapter := &authServiceAdapter{service: authService}

	authHandler := handler.NewAuth(logger, authAdapter)
	healthHandler := handler.NewHealth()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("GET /health", healthHandler.Health)

	srv := &http.Server{
		Addr:    ":" + appCfg.Port,
		Handler: mux,
	}

	go func() {
		logger.Info("Starting server on :" + appCfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", slog.String("err", err.Error()))
	}

	logger.Info("Server exited")
}
