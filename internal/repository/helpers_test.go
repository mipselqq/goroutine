package repository_test

import (
	"context"
	"testing"
	"time"

	"goroutine/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

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
