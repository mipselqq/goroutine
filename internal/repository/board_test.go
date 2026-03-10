//go:build integration

package repository_test

import (
	"context"
	"testing"

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

		const query = `SELECT owner_id, name, description FROM boards WHERE owner_id=$1 AND name=$2 AND description=$3`
		var dbOwnerID string
		var dbName string
		var dbDescription string
		err = pool.QueryRow(
			context.Background(),
			query,
			userID.String(),
			boardName.String(),
			boardDescription.String(),
		).Scan(&dbOwnerID, &dbName, &dbDescription)
		if err != nil {
			t.Errorf("Failed to find board in DB: %v", err)
		}
		if dbOwnerID != userID.String() {
			t.Errorf("Expected owner ID %q, got %q", userID.String(), dbOwnerID)
		}
		if dbName != boardName.String() {
			t.Errorf("Expected name %q, got %q", boardName.String(), dbName)
		}
		if dbDescription != boardDescription.String() {
			t.Errorf("Expected description %q, got %q", boardDescription.String(), dbDescription)
		}
		if board.CreatedAt != board.UpdatedAt {
			t.Errorf("Expected created at and updated at to be the same, got %v and %v", board.CreatedAt, board.UpdatedAt)
		}
	})
}
