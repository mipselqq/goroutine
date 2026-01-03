package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"go-todo/internal/config"

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

	goose.SetLogger(&gooseLogger{logger: logger})

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

type gooseLogger struct {
	logger *slog.Logger
}

func (l *gooseLogger) Fatal(v ...interface{}) {
	l.logger.Error(fmt.Sprint(v...))
	os.Exit(1)
}

func (l *gooseLogger) Fatalf(format string, v ...interface{}) {
	l.logger.Error(fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (l *gooseLogger) Print(v ...interface{}) {
	l.logger.Info(fmt.Sprint(v...))
}

func (l *gooseLogger) Printf(format string, v ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, v...))
}
