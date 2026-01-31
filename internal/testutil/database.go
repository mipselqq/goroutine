package testutil

import (
	"context"
	"testing"
	"time"

	"go-todo/internal/app"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func SetupTestDB(t *testing.T, migrationsDir string) *pgxpool.Pool {
	t.Helper()
	_ = godotenv.Load("../../.env.dev")
	logger := CreateTestLogger(t)

	pool, err := app.SetupDatabaseFromEnv(logger, migrationsDir)
	if err != nil {
		t.Fatalf("Failed to setup database: %v", err)
	}

	return pool
}

func TruncateTable(t *testing.T, pool *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx, "TRUNCATE TABLE users CASCADE")
	if err != nil {
		t.Fatalf("Failed to TRUNCATE TABLE users: %v", err)
	}
}
