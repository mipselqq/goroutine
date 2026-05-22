// Main application binary loading configuration and starting the servers.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"goroutine/docs/openapi"
	"goroutine/internal/app"
	"goroutine/internal/config"
	"goroutine/internal/http/httpschema"
	"goroutine/internal/logging"

	"github.com/prometheus/client_golang/prometheus"
)

var version = "no version bundled by linker"

// @title Goroutine kanban API
// @description A nice kanban board with a beautiful heart ✨
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter JWT token with `Bearer ` prefix, e.g. `Bearer eyJhbGciOi...`
func main() {
	if os.Getenv("ENV") != "prod" {
		_ = godotenv.Load(".env.dev")
	}

	bootLogger := logging.NewLogger("dev", "info")
	appCfg := config.NewAppConfigFromEnv(bootLogger)
	openapi.SwaggerInfo.Host = appCfg.SwaggerHost
	logger := logging.NewLogger(appCfg.Env, appCfg.LogLevel, httpschema.AllExtractors()...)

	logger.Info("Running", slog.String("version", version))
	logger.Info("App config", slog.Any("config", appCfg))

	pool, err := app.SetupPostgresFromEnv(logger, "migrations")
	if err != nil {
		os.Exit(1)
	}

	redisClient, err := app.SetupRedisFromEnv(logger)
	if err != nil {
		pool.Close()
		os.Exit(1)
	}

	defer pool.Close()
	defer func() {
		_ = redisClient.Close()
	}()

	application := app.New(logger, pool, redisClient, &appCfg, prometheus.DefaultRegisterer)

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
