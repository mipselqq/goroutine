//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"goroutine/internal/repository"
)

func TestBoardRepository_Create(t *testing.T) {
	pool := SetupTestDB(t, "../../migrations")
	defer pool.Close()

	r := repository.NewPgBoard(pool)

	t.Run("Success", func(t *testing.T) {
		TruncateTable(t, pool, "boards")
		TruncateTable(t, pool, "users")

		CreateUser(t, pool, userID, "test@example.com")

		board, err := r.Create(context.Background(), userID, boardName, boardDescription)
		if err != nil {
			t.Errorf("Create() error = %v", err)
		}
		if board.ID.IsEmpty() {
			t.Errorf("Expected board ID to be generated, got empty")
		}
		if board.OwnerID != userID {
			t.Errorf("Expected owner ID %q, got %q", userID, board.OwnerID)
		}
		if board.Name != boardName {
			t.Errorf("Expected name %q, got %q", boardName, board.Name)
		}
		if board.Description != boardDescription {
			t.Errorf("Expected description %q, got %q", boardDescription, board.Description)
		}
		if board.CreatedAt.IsZero() {
			t.Errorf("Expected created at to be set, got zero value")
		}
		if board.UpdatedAt.IsZero() {
			t.Errorf("Expected updated at to be set, got zero value")
		}
		if board.CreatedAt != board.UpdatedAt {
			t.Errorf("Expected created at and updated at to be the same, got %v and %v", board.CreatedAt, board.UpdatedAt)
		}

		const query = `
		SELECT owner_id, name, description, created_at, updated_at 
		FROM boards 
		WHERE id = $1`

		var (
			dbOwnerID     string
			dbName        string
			dbDescription string
			dbCreatedAt   time.Time
			dbUpdatedAt   time.Time
		)
		err = pool.QueryRow(context.Background(), query, board.ID.String()).
			Scan(&dbOwnerID, &dbName, &dbDescription, &dbCreatedAt, &dbUpdatedAt)
		if err != nil {
			t.Fatalf("Failed to find board in DB by ID %q: %v", board.ID, err)
		}
		if dbOwnerID != userID.String() {
			t.Errorf("DB: expected owner ID %q, got %q", userID, dbOwnerID)
		}
		if dbName != boardName.String() {
			t.Errorf("DB: expected name %q, got %q", boardName, dbName)
		}
		if dbDescription != boardDescription.String() {
			t.Errorf("DB: expected description %q, got %q", boardDescription, dbDescription)
		}
		if !dbCreatedAt.Equal(board.CreatedAt) {
			t.Errorf("DB: expected created_at %v, got %v", board.CreatedAt, dbCreatedAt)
		}
		if !dbUpdatedAt.Equal(board.UpdatedAt) {
			t.Errorf("DB: expected updated_at %v, got %v", board.UpdatedAt, dbUpdatedAt)
		}
	})
}
