package testutil

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"goroutine/internal/app"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func SetupTestDB(t *testing.T, migrationsDir string) *pgxpool.Pool {
	t.Helper()
	_ = godotenv.Load("../../.env.dev", "../.env.dev")
	logger := NewTestLogger(t)

	pool, err := app.SetupDatabaseFromEnv(logger, migrationsDir)
	if err != nil {
		t.Fatalf("SetupDatabaseFromEnv() error = %v", err)
	}

	return pool
}

func TruncateTable(t *testing.T, pool *pgxpool.Pool, name string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", pgx.Identifier{name}.Sanitize())

	_, err := pool.Exec(ctx, query)
	if err != nil {
		t.Fatalf("TRUNCATE TABLE %q error = %v", name, err)
	}
}

// TruncateAllTables clears application tables in dependency order (FK-safe).
func TruncateAllTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	names := []string{"columns", "boards", "users"}
	parts := make([]string, len(names))
	for i, name := range names {
		parts[i] = pgx.Identifier{name}.Sanitize()
	}
	query := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", strings.Join(parts, ", "))

	_, err := pool.Exec(ctx, query)
	if err != nil {
		t.Fatalf("TRUNCATE ALL application tables error = %v", err)
	}
}
