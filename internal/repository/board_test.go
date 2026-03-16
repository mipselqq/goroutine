//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/testutil"
)

func TestBoardRepository_Create(t *testing.T) {
	pool := testutil.SetupTestDB(t, "../../migrations")
	defer pool.Close()

	r := repository.NewPgBoard(pool)
	userID := testutil.ValidUserID()
	boardName := testutil.ValidBoardName()
	boardDescription := testutil.ValidBoardDescription()

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "boards")
		testutil.TruncateTable(t, pool, "users")

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
		if !board.CreatedAt.Equal(board.UpdatedAt) {
			t.Errorf("Expected created at and updated at to be the same, got %v and %v", board.CreatedAt, board.UpdatedAt)
		}

		const query = `
		SELECT owner_id, name, description, created_at, updated_at 
		FROM boards 
		WHERE id = $1`

		var (
			dbOwnerID     domain.UserID
			dbName        domain.BoardName
			dbDescription domain.BoardDescription
			dbCreatedAt   time.Time
			dbUpdatedAt   time.Time
		)
		err = pool.QueryRow(context.Background(), query, board.ID).
			Scan(&dbOwnerID, &dbName, &dbDescription, &dbCreatedAt, &dbUpdatedAt)
		if err != nil {
			t.Fatalf("Failed to find board in DB by ID %q: %v", board.ID, err)
		}
		if dbOwnerID != userID {
			t.Errorf("DB: expected owner ID %q, got %q", userID, dbOwnerID)
		}
		if dbName != boardName {
			t.Errorf("DB: expected name %q, got %q", boardName, dbName)
		}
		if dbDescription != boardDescription {
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
