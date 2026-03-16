//go:build e2e

package tests

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"goroutine/internal/app"
	"goroutine/internal/config"
	"goroutine/internal/testutil"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
)

func E2EPrelude(t *testing.T) (*http.Client, *httptest.Server, *pgxpool.Pool) {
	t.Helper()

	pool := testutil.SetupTestDB(t, "../migrations")

	logger := testutil.NewTestLogger(t)
	cfg := config.NewAppConfigFromEnv(logger)
	logger.Info("App config", slog.Any("config", cfg))

	application := app.New(logger, pool, &cfg, prometheus.NewRegistry())

	ts := httptest.NewServer(application.Router)
	t.Cleanup(func() {
		ts.Close()
		pool.Close()
	})

	client := ts.Client()

	return client, ts, pool
}
