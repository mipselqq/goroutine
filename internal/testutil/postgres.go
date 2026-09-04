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
)

func SetupPostgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	mustLoadDevEnv()
	logger := NewLogger(t)

	pool, err := app.SetupPostgresFromEnv(logger)
	if err != nil {
		t.Fatalf("Failed to setup Postgres: %v", err)
	}

	return pool
}

// TruncateAllTables clears application tables in dependency order (FK-safe).
func TruncateAllTables(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	names := []string{"tasks", "columns", "boards", "users", "notification_outbox"}
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
