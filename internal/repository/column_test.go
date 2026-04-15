//go:build integration

package repository_test

import (
	"context"
	"errors"
	"reflect"
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

func TestColumnRepository_ListByBoardID(t *testing.T) {
	pool := testutil.SetupTestDB(t, "../../migrations")
	defer pool.Close()

	r := repository.NewPgColumn(pool)

	t.Run("Success empty", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "columns")
		testutil.TruncateTable(t, pool, "boards")
		testutil.TruncateTable(t, pool, "users")

		userID := testutil.ValidUserID()
		CreateUser(t, pool, userID, "column-list-empty@example.com")

		board := testutil.ValidBoard()
		InsertBoard(t, pool, &board)

		columns, err := r.ListByBoardID(context.Background(), board.ID)
		if err != nil {
			t.Fatalf("ListByBoardID() error = %v", err)
		}
		if len(columns) != 0 {
			t.Fatalf("expected 0 columns, got %d", len(columns))
		}
	})

	t.Run("Success ordered and filtered by board", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "columns")
		testutil.TruncateTable(t, pool, "boards")
		testutil.TruncateTable(t, pool, "users")

		userID := testutil.ValidUserID()
		CreateUser(t, pool, userID, "column-list-ordered@example.com")

		boardA := testutil.ValidBoard()
		InsertBoard(t, pool, &boardA)

		boardB := testutil.ValidBoard()
		InsertBoard(t, pool, &boardB)

		name1, _ := domain.NewColumnName("Todo")
		name2, _ := domain.NewColumnName("In Progress")
		name3, _ := domain.NewColumnName("Done")
		now := testutil.FixedTimeNow()

		first, err := r.Create(context.Background(), boardA.ID, name1, now, now)
		if err != nil {
			t.Fatalf("Create first: %v", err)
		}
		second, err := r.Create(context.Background(), boardA.ID, name2, now, now)
		if err != nil {
			t.Fatalf("Create second: %v", err)
		}
		_, err = r.Create(context.Background(), boardB.ID, name3, now, now)
		if err != nil {
			t.Fatalf("Create other board: %v", err)
		}

		got, err := r.ListByBoardID(context.Background(), boardA.ID)
		if err != nil {
			t.Fatalf("ListByBoardID() error = %v", err)
		}

		want := []domain.Column{first, second}
		if !reflect.DeepEqual(want, got) {
			t.Errorf("ListByBoardID() = %#v, want %#v", got, want)
		}
	})
}

func TestColumnRepository_GetByID(t *testing.T) {
	pool := testutil.SetupTestDB(t, "../../migrations")
	defer pool.Close()

	r := repository.NewPgColumn(pool)

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "columns")
		testutil.TruncateTable(t, pool, "boards")
		testutil.TruncateTable(t, pool, "users")

		userID := testutil.ValidUserID()
		CreateUser(t, pool, userID, "column-getbyid@example.com")

		board := testutil.ValidBoard()
		InsertBoard(t, pool, &board)

		name, _ := domain.NewColumnName("Todo")
		now := testutil.FixedTimeNow()
		created, err := r.Create(context.Background(), board.ID, name, now, now)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		got, err := r.GetByID(context.Background(), created.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if !reflect.DeepEqual(created, got) {
			t.Errorf("GetByID() = %#v, want %#v", got, created)
		}
	})

	t.Run("Not found", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "columns")

		_, err := r.GetByID(context.Background(), domain.NewColumnID())
		if !errors.Is(err, repository.ErrRowNotFound) {
			t.Errorf("GetByID() error = %v, want ErrRowNotFound", err)
		}
	})
}

func TestColumnRepository_UpdateByID(t *testing.T) {
	pool := testutil.SetupTestDB(t, "../../migrations")
	defer pool.Close()

	r := repository.NewPgColumn(pool)

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "columns")
		testutil.TruncateTable(t, pool, "boards")
		testutil.TruncateTable(t, pool, "users")

		userID := testutil.ValidUserID()
		CreateUser(t, pool, userID, "column-update@example.com")

		board := testutil.ValidBoard()
		InsertBoard(t, pool, &board)

		name, _ := domain.NewColumnName("Todo")
		now := testutil.FixedTimeNow()
		created, err := r.Create(context.Background(), board.ID, name, now, now)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		updatedName, _ := domain.NewColumnName("Renamed")
		updatedAt := testutil.FixedTime5mFromNow()
		updated, err := r.UpdateByID(context.Background(), board.ID, created.ID, &updatedName, updatedAt)
		if err != nil {
			t.Fatalf("UpdateByID() error = %v", err)
		}

		if updated.Name != updatedName {
			t.Errorf("expected name %q, got %q", updatedName, updated.Name)
		}
		if !updated.UpdatedAt.Equal(updatedAt) {
			t.Errorf("expected updatedAt %v, got %v", updatedAt, updated.UpdatedAt)
		}
		if updated.Position != created.Position {
			t.Errorf("expected position %v, got %v", created.Position, updated.Position)
		}
	})

	t.Run("Not found by column id", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "columns")
		testutil.TruncateTable(t, pool, "boards")
		testutil.TruncateTable(t, pool, "users")

		userID := testutil.ValidUserID()
		CreateUser(t, pool, userID, "column-update-missing-col@example.com")

		board := testutil.ValidBoard()
		InsertBoard(t, pool, &board)

		updatedName, _ := domain.NewColumnName("Renamed")
		_, err := r.UpdateByID(context.Background(), board.ID, domain.NewColumnID(), &updatedName, testutil.FixedTime5mFromNow())
		if !errors.Is(err, repository.ErrRowNotFound) {
			t.Errorf("UpdateByID() error = %v, want ErrRowNotFound", err)
		}
	})

	t.Run("Not found by board id", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "columns")
		testutil.TruncateTable(t, pool, "boards")
		testutil.TruncateTable(t, pool, "users")

		userID := testutil.ValidUserID()
		CreateUser(t, pool, userID, "column-update-missing-board@example.com")

		board := testutil.ValidBoard()
		InsertBoard(t, pool, &board)

		name, _ := domain.NewColumnName("Todo")
		created, err := r.Create(context.Background(), board.ID, name, testutil.FixedTimeNow(), testutil.FixedTimeNow())
		if err != nil {
			t.Fatalf("Create: %v", err)
		}

		updatedName, _ := domain.NewColumnName("Renamed")
		_, err = r.UpdateByID(context.Background(), domain.NewBoardID(), created.ID, &updatedName, testutil.FixedTime5mFromNow())
		if !errors.Is(err, repository.ErrRowNotFound) {
			t.Errorf("UpdateByID() error = %v, want ErrRowNotFound", err)
		}
	})
}
