package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"goroutine/internal/config"
	"goroutine/internal/logging"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func SetupDatabaseFromEnv(logger *slog.Logger, migrationsDir string) (*pgxpool.Pool, error) {
	cfg := config.NewPGConfigFromEnv()

	logger.Info("Database config", slog.Any("config", cfg))

	poolConfig, err := cfg.ParsePGXpoolConfig()
	if err != nil {
		logger.Error("Failed to parse database config", slog.String("err", err.Error()))
		return nil, err
	}

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		logger.Error("Failed to connect to database", slog.String("err", err.Error()))
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		logger.Error("Failed to ping database", slog.String("err", err.Error()))
		return nil, err
	}

	goose.SetLogger(&logging.GooseLogger{Logger: logger})

	if err := goose.SetDialect("postgres"); err != nil {
		logger.Error("Failed to set goose dialect", slog.String("err", err.Error()))
		return nil, err
	}

	db := stdlib.OpenDBFromPool(pool)

	if err := goose.Up(db, migrationsDir); err != nil {
		logger.Error("Failed to run migrations", slog.String("err", err.Error()))
		return nil, err
	}

	return pool, nil
}

func RunBackgroundServer(logger *slog.Logger, name string, addr string, handler http.Handler) *http.Server {
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		logger.Info("Starting "+name+" on http://"+addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(name+" failed", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}()

	return srv
}
