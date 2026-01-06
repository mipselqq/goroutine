package testutil

import (
	"log/slog"
	"testing"

	"go-todo/internal/app"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func SetupTestDB(t *testing.T) (*pgxpool.Pool, *slog.Logger) {
	_ = godotenv.Load("../../.env.dev")
	logger := CreateTestLogger(t)

	pool, err := app.SetupDatabaseFromEnv(logger, "../../migrations")
	if err != nil {
		t.Fatalf("Failed to setup database: %v", err)
	}

	return pool, logger
}
