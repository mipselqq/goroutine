package app

import (
	"context"
	"fmt"
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
	logger = logging.WithModule(logger, "app.startup")

	envConfig := config.NewPGFromEnv(logger)

	logger.Info("Database config", slog.Any("config", envConfig))

	cfg, err := pgxpool.ParseConfig(envConfig.BuildDSN().RevealSecret())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %v", err)
	}

	pingTimeout := 5 * time.Second
	cfg.ConnConfig.ConnectTimeout = 10 * time.Second
	cfg.PingTimeout = pingTimeout
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 5 * time.Minute

	if cfg.ConnConfig.RuntimeParams == nil {
		cfg.ConnConfig.RuntimeParams = map[string]string{}
	}
	cfg.ConnConfig.RuntimeParams["statement_timeout"] = "5s"
	cfg.ConnConfig.RuntimeParams["timezone"] = "UTC"

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	err = pool.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	goose.SetLogger(&logging.GooseLogger{Base: logger})

	err = goose.SetDialect("postgres")
	if err != nil {
		return nil, fmt.Errorf("failed to set goose dialect: %v", err)
	}

	db := stdlib.OpenDBFromPool(pool)

	err = goose.Up(db, migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	return pool, nil
}

func SetupRedisFromEnv(logger *slog.Logger) (*redis.Client, error) {
	logger = logging.WithModule(logger, "app.startup")

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

	err := client.Ping(context.Background()).Err() // Uses client timeouts
	if err != nil {
		return nil, fmt.Errorf("failed to ping redis: %v", err)
	}

	return client, nil
}

func RunBackgroundServer(logger *slog.Logger, name, addr string, handler http.Handler) *http.Server {
	logger = logging.WithModule(logger, "app.startup")

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
