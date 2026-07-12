package app

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"goroutine/internal/config"
	"goroutine/internal/logging"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
)

func SetupPostgresFromEnv(logger *slog.Logger, migrationsDir string) (*pgxpool.Pool, error) {
	envConfig := config.NewPGFromEnv(logger)

	logger.Info("Database config", slog.Any("config", envConfig))

	cfg, err := pgxpool.ParseConfig(envConfig.BuildDSN().RevealSecret())
	if err != nil {
		logger.Error("Failed to parse database config", slog.String("err", err.Error()))
		return nil, err
	}

	cfg.ConnConfig.ConnectTimeout = 10 * time.Second
	cfg.PingTimeout = 5 * time.Second
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute

	if cfg.ConnConfig.RuntimeParams == nil {
		cfg.ConnConfig.RuntimeParams = map[string]string{}
	}
	cfg.ConnConfig.RuntimeParams["statement_timeout"] = "5s"
	cfg.ConnConfig.RuntimeParams["timezone"] = "UTC"

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		logger.Error("Failed to connect to database", slog.String("err", err.Error()))
		return nil, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		logger.Error("Failed to ping database", slog.String("err", err.Error()))
		return nil, err
	}

	goose.SetLogger(&logging.GooseLogger{Base: logger})

	err = goose.SetDialect("postgres")
	if err != nil {
		logger.Error("Failed to set goose dialect", slog.String("err", err.Error()))
		return nil, err
	}

	db := stdlib.OpenDBFromPool(pool)

	err = goose.Up(db, migrationsDir)
	if err != nil {
		logger.Error("Failed to run migrations", slog.String("err", err.Error()))
		return nil, err
	}

	return pool, nil
}

func SetupRedisFromEnv(logger *slog.Logger) (*redis.Client, error) {
	cfg := config.NewRedisFromEnv(logger)

	logger.Info("Redis config", slog.Any("config", cfg))

	client := redis.NewClient(&redis.Options{
		Addr:            cfg.BuildAddr(),
		Password:        cfg.Password.RevealSecret(),
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
		PoolTimeout:     4 * time.Second,
		ConnMaxIdleTime: 5 * time.Minute,
	})

	err := client.Ping(context.Background()).Err()
	if err != nil {
		logger.Error("Failed to ping Redis", slog.String("err", err.Error()))
		return nil, err
	}

	return client, nil
}

func RunBackgroundServer(logger *slog.Logger, name, addr string, handler http.Handler) *http.Server {
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Info("Starting " + name + " on http://" + addr)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Error(name+" failed", slog.String("err", err.Error()))
			os.Exit(1)
		}
	}()

	return srv
}
