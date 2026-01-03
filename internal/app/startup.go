package app

import (
	"context"
	"log/slog"

	"go-todo/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func SetupDatabaseFromEnv(logger *slog.Logger) (*pgxpool.Pool, error) {
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

	return pool, nil
}
