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
	httpSwagger "github.com/swaggo/http-swagger/v2"

	"go-todo/internal/app"
	"go-todo/internal/config"
	"go-todo/internal/handler"
	"go-todo/internal/logging"
	"go-todo/internal/repository"
	"go-todo/internal/service"

	_ "go-todo/docs"
)

// @title Go Todo API
// @version 1.0
// @description A todo project for learning Go-go-go-go
// @host localhost:8080
// @BasePath /
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
	authService := service.NewAuth(userRepo, service.JWTOptions{
		JWTSecret: appCfg.JWTSecret,
		Exp:       appCfg.JWTExp,
	})

	authHandler := handler.NewAuth(logger, authService)
	healthHandler := handler.NewHealth()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /register", authHandler.Register)
	mux.HandleFunc("POST /login", authHandler.Login)
	mux.HandleFunc("GET /health", healthHandler.Health)
	// TODO: restrict unpriviliged access
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)

	addr := appCfg.Host + ":" + appCfg.Port
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		logger.Info("Starting server on http://" + addr)
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
