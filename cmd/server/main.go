package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"go-todo/internal/app"
	"go-todo/internal/config"
	"go-todo/internal/domain"
	"go-todo/internal/handler"
	"go-todo/internal/logging"
	"go-todo/internal/repository"
	"go-todo/internal/service"
)

type authServiceAdapter struct {
	service *service.Auth
}

func (a *authServiceAdapter) Register(ctx context.Context, email domain.Email, password domain.Password) (string, error) {
	if err := a.service.Register(ctx, email, password); err != nil {
		return "", err
	}

	return "simulated_token", nil
}

func main() {
	if os.Getenv("ENV") != "prod" {
		_ = godotenv.Load(".env.dev")
	}

	appCfg := config.NewAppConfigFromEnv()
	logger := logging.NewLogger(appCfg.Env, appCfg.LogLevel)

	logger.Info("App config", slog.Any("config", appCfg))

	pool, err := app.SetupDatabaseFromEnv(logger, "migrations")
	if err != nil {
		os.Exit(1)
	}
	defer pool.Close()

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
