package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"goroutine/docs"
	"goroutine/internal/app"
	"goroutine/internal/config"
	"goroutine/internal/logging"
)

var version = "no version bundled by linker"

// @title Goroutine kanban API
// @description A nice kanban board with a beautiful heart ✨
// @BasePath /
func main() {
	if os.Getenv("ENV") != "prod" {
		_ = godotenv.Load(".env.dev")
	}

	bootLogger := logging.NewTintStdoutLogger("info")
	appCfg := config.NewAppConfigFromEnv(bootLogger)
	docs.SwaggerInfo.Host = appCfg.SwaggerHost
	logger := logging.NewLogger(appCfg.Env, appCfg.LogLevel)

	logger.Info("Running", slog.String("version", version))
	logger.Info("App config", slog.Any("config", appCfg))

	pool, err := app.SetupDatabaseFromEnv(logger, "migrations")
	if err != nil {
		os.Exit(1)
	}
	defer pool.Close()

	application := app.New(logger, pool, &appCfg)

	srv := app.RunBackgroundServer(logger, "server", appCfg.Host+":"+appCfg.Port, application.Router)
	adminSrv := app.RunBackgroundServer(logger, "admin server", appCfg.Host+":"+appCfg.AdminPort, application.AdminRouter)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("Shutting down servers...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", slog.String("err", err.Error()))
	}

	if err := adminSrv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Admin server forced to shutdown", slog.String("err", err.Error()))
	}

	logger.Info("Server exited")
}
