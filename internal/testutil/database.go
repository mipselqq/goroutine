package testutil

import (
	"context"
	"testing"
	"time"

	"go-todo/internal/app"
	"go-todo/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func SetupTestDB(t *testing.T) *pgxpool.Pool {
	_ = godotenv.Load("../../.env.dev")
	logger := CreateTestLogger(t)

	pool, err := app.SetupDatabaseFromEnv(logger, "../../migrations")
	if err != nil {
		t.Fatalf("Failed to setup database: %v", err)
	}

	return pool
}

func SetupUserRepository(t *testing.T) (*repository.PgUser, *pgxpool.Pool) {
	t.Helper()
	pool := SetupTestDB(t)
	r := repository.NewPgUser(pool)

	return r, pool
}

func TruncateTable(t *testing.T, pool *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx, "TRUNCATE TABLE users CASCADE")
	if err != nil {
		t.Fatalf("Failed to TRUNCATE TABLE users: %v", err)
	}
}
