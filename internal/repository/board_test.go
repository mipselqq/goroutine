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

		err := r.Create(context.Background(), userID, boardName, boardDescription)
		if err != nil {
			t.Errorf("Create() error = %v", err)
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
	})
}
