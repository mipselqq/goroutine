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
	_, err := pool.Exec(ctx, query, id, email, "hash")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
}

// InsertBoard writes a board row directly; do not use the repository under test.
func InsertBoard(t *testing.T, pool *pgxpool.Pool, board *domain.Board) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const q = `
		INSERT INTO boards (id, owner_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := pool.Exec(ctx, q,
		board.ID, board.OwnerID, board.Name, board.Description, board.CreatedAt, board.UpdatedAt)
	if err != nil {
		t.Fatalf("insert board: %v", err)
	}
}
