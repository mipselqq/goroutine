package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"goroutine/internal/app"
	"goroutine/internal/domain"
	"goroutine/internal/testutil"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var (
	userID              = testutil.ParseUserID("018e1000-0000-7000-8000-000000000000")
	boardName, _        = domain.NewBoardName("Test Board")
	boardDescription, _ = domain.NewBoardDescription("Test Board Description")
)

// TODO: write local repository/helpers_test.go and use it here
// to avoid uncontrolled global testutil growth
func SetupTestDB(t *testing.T, migrationsDir string) *pgxpool.Pool {
	t.Helper()
	_ = godotenv.Load("../../.env.dev")
	logger := testutil.NewTestLogger(t)

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
		t.Fatalf("Failed to TRUNCATE TABLE %q: %v", name, err)
	}
}

func CreateUser(t *testing.T, pool *pgxpool.Pool, id domain.UserID, email string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const query = `INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)`
	_, err := pool.Exec(ctx, query, id.String(), email, "hash")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
}
