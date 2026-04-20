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

		validColumn := testutil.ValidColumn(board.ID)

		column, err := r.Create(
			context.Background(),
			board.ID,
			validColumn.Name,
			validColumn.CreatedAt,
			validColumn.UpdatedAt,
		)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if column.ID.IsEmpty() {
			t.Error("expected generated column id")
		}
		if column.BoardID != board.ID {
			t.Errorf("expected boardID %q, got %q", board.ID, column.BoardID)
		}
		if column.Name != validColumn.Name {
			t.Errorf("expected name %q, got %q", validColumn.Name, column.Name)
		}
		if column.Position.Int64() != 1 {
			t.Errorf("expected position 1, got %d", column.Position.Int64())
		}
		if !column.CreatedAt.Equal(validColumn.CreatedAt) {
			t.Errorf("expected createdAt %v, got %v", validColumn.CreatedAt, column.CreatedAt)
		}
		if !column.UpdatedAt.Equal(validColumn.UpdatedAt) {
			t.Errorf("expected updatedAt %v, got %v", validColumn.UpdatedAt, column.UpdatedAt)
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

	existing := testutil.ValidColumn(board.ID)
	InsertColumn(t, pool, &existing)

	toCreate := testutil.ValidColumn(board.ID)
	toCreate = testutil.UpdateValidColumn(t, &toCreate, "Done", toCreate.UpdatedAt)

	second, err := r.Create(
		context.Background(),
		board.ID,
		toCreate.Name,
		toCreate.CreatedAt,
		toCreate.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("Create second: %v", err)
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

		first := testutil.ValidColumn(boardA.ID)
		second := testutil.NewValidColumn(t, boardA.ID, "In Progress", 2)
		otherBoardColumn := testutil.NewValidColumn(t, boardB.ID, "Done", 1)

		InsertColumn(t, pool, &first)
		InsertColumn(t, pool, &second)
		InsertColumn(t, pool, &otherBoardColumn)

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

		created := testutil.ValidColumn(board.ID)
		InsertColumn(t, pool, &created)

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

		created := testutil.ValidColumn(board.ID)
		InsertColumn(t, pool, &created)

		want := testutil.UpdateValidColumn(t, &created, "Renamed", testutil.FixedTime5mFromNow())
		updated, err := r.UpdateByID(context.Background(), board.ID, created.ID, &want.Name, want.UpdatedAt)
		if err != nil {
			t.Fatalf("UpdateByID() error = %v", err)
		}

		if !reflect.DeepEqual(want, updated) {
			t.Errorf("UpdateByID() = %#v, want %#v", updated, want)
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

		created := testutil.ValidColumn(board.ID)
		InsertColumn(t, pool, &created)

		want := testutil.UpdateValidColumn(t, &created, "Renamed", testutil.FixedTime5mFromNow())
		_, err := r.UpdateByID(context.Background(), domain.NewBoardID(), created.ID, &want.Name, want.UpdatedAt)
		if !errors.Is(err, repository.ErrRowNotFound) {
			t.Errorf("UpdateByID() error = %v, want ErrRowNotFound", err)
		}
	})
}

func TestColumnRepository_Delete(t *testing.T) {
	pool := testutil.SetupTestDB(t, "../../migrations")
	defer pool.Close()

	r := repository.NewPgColumn(pool)

	t.Run("Success shift positions", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "columns")
		testutil.TruncateTable(t, pool, "boards")
		testutil.TruncateTable(t, pool, "users")

		userID := testutil.ValidUserID()
		CreateUser(t, pool, userID, "column-delete-shift@example.com")

		board := testutil.ValidBoard()
		InsertBoard(t, pool, &board)

		first := testutil.ValidColumn(board.ID)
		second := testutil.NewValidColumn(t, board.ID, "In Progress", 2)
		third := testutil.NewValidColumn(t, board.ID, "Done", 3)

		InsertColumn(t, pool, &first)
		InsertColumn(t, pool, &second)
		InsertColumn(t, pool, &third)

		err := r.Delete(context.Background(), board.ID, second.ID)
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		got := ListColumnsByBoardID(t, pool, board.ID)

		if len(got) != 2 {
			t.Fatalf("expected 2 columns after delete, got %d", len(got))
		}
		if got[0].ID != first.ID {
			t.Errorf("expected first column id %q, got %q", first.ID, got[0].ID)
		}
		if got[0].Position.Int64() != 1 {
			t.Errorf("expected first position 1, got %d", got[0].Position.Int64())
		}
		if got[1].ID != third.ID {
			t.Errorf("expected second column id %q, got %q", third.ID, got[1].ID)
		}
		if got[1].Position.Int64() != 2 {
			t.Errorf("expected second position 2 after shift, got %d", got[1].Position.Int64())
		}

		_, ok := FindColumnByID(t, pool, second.ID)
		if ok {
			t.Error("expected deleted column to be absent in DB")
		}
	})

	t.Run("Not found by column id", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "columns")
		testutil.TruncateTable(t, pool, "boards")
		testutil.TruncateTable(t, pool, "users")

		userID := testutil.ValidUserID()
		CreateUser(t, pool, userID, "column-delete-missing-col@example.com")

		board := testutil.ValidBoard()
		InsertBoard(t, pool, &board)

		err := r.Delete(context.Background(), board.ID, domain.NewColumnID())
		if !errors.Is(err, repository.ErrRowNotFound) {
			t.Errorf("Delete() error = %v, want ErrRowNotFound", err)
		}
	})

	t.Run("Not found by board id", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "columns")
		testutil.TruncateTable(t, pool, "boards")
		testutil.TruncateTable(t, pool, "users")

		userID := testutil.ValidUserID()
		CreateUser(t, pool, userID, "column-delete-missing-board@example.com")

		board := testutil.ValidBoard()
		InsertBoard(t, pool, &board)

		created := testutil.ValidColumn(board.ID)
		InsertColumn(t, pool, &created)

		err := r.Delete(context.Background(), domain.NewBoardID(), created.ID)
		if !errors.Is(err, repository.ErrRowNotFound) {
			t.Errorf("Delete() error = %v, want ErrRowNotFound", err)
		}
	})
}
