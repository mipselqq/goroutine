package testutil

import (
	"context"
	"fmt"
	"testing"
	"time"

	app "goroutine/internal"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func SetupTestDB(t *testing.T, migrationsDir string) *pgxpool.Pool {
	t.Helper()
	_ = godotenv.Load("../../.env.dev")
	logger := NewTestLogger(t)

	pool, err := app.SetupDatabaseFromEnv(logger, migrationsDir)
	if err != nil {
		t.Fatalf("Failed to setup database: %v", err)
	}

	return pool
}

func TruncateTable(t *testing.T, pool *pgxpool.Pool, name string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", pgx.Identifier{name}.Sanitize())

	_, err := pool.Exec(ctx, query)
	if err != nil {
		t.Fatalf("Failed to TRUNCATE TABLE %s: %v", name, err)
	}
}
