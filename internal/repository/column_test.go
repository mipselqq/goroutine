//go:build integration

package repository_test

import (
	"context"
	"testing"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/testutil"
)

func TestColumnRepository_Create(t *testing.T) {
	pool := testutil.SetupTestDB(t, "../../migrations")
	defer pool.Close()

	r := repository.NewPgColumn(pool)

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "columns")
		testutil.TruncateTable(t, pool, "boards")
		testutil.TruncateTable(t, pool, "users")

		userID := testutil.ValidUserID()
		CreateUser(t, pool, userID, "column-create@example.com")

		board := testutil.ValidBoard()
		board.OwnerID = userID
		InsertBoard(t, pool, &board)

		name, err := domain.NewColumnName("In Progress")
		if err != nil {
			t.Fatalf("NewColumnName: %v", err)
		}
		now := testutil.FixedTimeNow()

		column, err := r.Create(context.Background(), board.ID, name, now, now)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if column.ID.IsEmpty() {
			t.Error("expected generated column id")
		}
		if column.BoardID != board.ID {
			t.Errorf("expected boardID %q, got %q", board.ID, column.BoardID)
		}
		if column.Name != name {
			t.Errorf("expected name %q, got %q", name, column.Name)
		}
		if column.Position.Int64() != 1 {
			t.Errorf("expected position 1, got %d", column.Position.Int64())
		}
		if !column.CreatedAt.Equal(now) {
			t.Errorf("expected createdAt %v, got %v", now, column.CreatedAt)
		}
		if !column.UpdatedAt.Equal(now) {
			t.Errorf("expected updatedAt %v, got %v", now, column.UpdatedAt)
		}
	})
}

func TestColumnRepository_Create_AppendsPosition(t *testing.T) {
	pool := testutil.SetupTestDB(t, "../../migrations")
	defer pool.Close()

	r := repository.NewPgColumn(pool)

	testutil.TruncateTable(t, pool, "columns")
	testutil.TruncateTable(t, pool, "boards")
	testutil.TruncateTable(t, pool, "users")

	userID := testutil.ValidUserID()
	CreateUser(t, pool, userID, "column-max@example.com")

	board := testutil.ValidBoard()
	board.OwnerID = userID
	InsertBoard(t, pool, &board)

	nameA, _ := domain.NewColumnName("Todo")
	nameB, _ := domain.NewColumnName("Done")
	now := testutil.FixedTimeNow()

	first, err := r.Create(context.Background(), board.ID, nameA, now, now)
	if err != nil {
		t.Fatalf("Create first: %v", err)
	}
	second, err := r.Create(context.Background(), board.ID, nameB, now, now)
	if err != nil {
		t.Fatalf("Create second: %v", err)
	}

	if first.Position.Int64() != 1 {
		t.Errorf("first position expected 1, got %d", first.Position.Int64())
	}
	if second.Position.Int64() != 2 {
		t.Errorf("second position expected 2, got %d", second.Position.Int64())
	}
}
