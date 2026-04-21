//go:build integration

package repository_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

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
		)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if column.ID.IsEmpty() {
			t.Error("got empty column id, want generated id")
		}
		if column.BoardID != board.ID {
			t.Errorf("got boardID %q, want %q", column.BoardID, board.ID)
		}
		if column.Name != validColumn.Name {
			t.Errorf("got name %q, want %q", column.Name, validColumn.Name)
		}
		if column.Position.Int64() != 1 {
			t.Errorf("got position %d, want 1", column.Position.Int64())
		}
		if column.CreatedAt.IsZero() {
			t.Errorf("got zero createdAt, want set value")
		}
		if column.UpdatedAt.IsZero() {
			t.Errorf("got zero updatedAt, want set value")
		}
		if !column.CreatedAt.Equal(column.UpdatedAt) {
			t.Errorf("got createdAt=%v updatedAt=%v, want equal", column.CreatedAt, column.UpdatedAt)
		}
		AssertTimestampPrecisionAtLeastMillis(t, pool, "columns", "created_at", "updated_at")

		stored, ok := FindColumnByID(t, pool, column.ID)
		if !ok {
			t.Fatalf("created column %q not found in DB", column.ID)
		}
		if !reflect.DeepEqual(column, stored) {
			t.Errorf("stored column = %#v, want %#v", stored, column)
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
	)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if second.Position.Int64() != 2 {
		t.Errorf("got second position %d, want 2", second.Position.Int64())
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
			t.Fatalf("got %d columns, want 0", len(columns))
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

	assertUpdatedColumn := func(t *testing.T, got domain.Column, want domain.Column) {
		t.Helper()

		if got.ID != want.ID {
			t.Errorf("got id %q, want %q", got.ID, want.ID)
		}
		if got.BoardID != want.BoardID {
			t.Errorf("got boardID %q, want %q", got.BoardID, want.BoardID)
		}
		if got.Name != want.Name {
			t.Errorf("got name %q, want %q", got.Name, want.Name)
		}
		if got.Position != want.Position {
			t.Errorf("got position %d, want %d", got.Position.Int64(), want.Position.Int64())
		}
		if !got.CreatedAt.Truncate(time.Millisecond).Equal(want.CreatedAt.Truncate(time.Millisecond)) {
			t.Errorf("got createdAt %v, want %v (at millisecond precision)", got.CreatedAt, want.CreatedAt)
		}
		if !got.UpdatedAt.After(want.UpdatedAt) {
			t.Errorf("got updatedAt %v, want after %v", got.UpdatedAt, want.UpdatedAt)
		}
		AssertTimestampPrecisionAtLeastMillis(t, pool, "columns", "created_at", "updated_at")

		stored, ok := FindColumnByID(t, pool, want.ID)
		if !ok {
			t.Fatalf("updated column %q not found in DB", want.ID)
		}
		if !reflect.DeepEqual(got, stored) {
			t.Errorf("stored column = %#v, want %#v", stored, got)
		}
	}

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "columns")
		testutil.TruncateTable(t, pool, "boards")
		testutil.TruncateTable(t, pool, "users")

		userID := testutil.ValidUserID()
		CreateUser(t, pool, userID, "column-update@example.com")

		board := testutil.ValidBoard()
		InsertBoard(t, pool, &board)

		created := testutil.ValidColumn(board.ID)
		createdAtBeforeUpdate := time.Now().UTC()
		updatedAtBeforeUpdate := createdAtBeforeUpdate
		created.CreatedAt = createdAtBeforeUpdate
		created.UpdatedAt = updatedAtBeforeUpdate
		InsertColumn(t, pool, &created)

		want := testutil.UpdateValidColumn(t, &created, "Renamed", testutil.FixedTime5mFromNow())
		updated, err := r.UpdateByID(context.Background(), board.ID, created.ID, &want.Name)
		if err != nil {
			t.Fatalf("UpdateByID() error = %v", err)
		}

		assertUpdatedColumn(t, updated, want)
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
		_, err := r.UpdateByID(context.Background(), board.ID, domain.NewColumnID(), &updatedName)
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
		_, err := r.UpdateByID(context.Background(), domain.NewBoardID(), created.ID, &want.Name)
		if !errors.Is(err, repository.ErrRowNotFound) {
			t.Errorf("UpdateByID() error = %v, want ErrRowNotFound", err)
		}
	})
}

func TestColumnRepository_Move(t *testing.T) {
	pool := testutil.SetupTestDB(t, "../../migrations")
	defer pool.Close()

	r := repository.NewPgColumn(pool)

	t.Run("Success move down", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "columns")
		testutil.TruncateTable(t, pool, "boards")
		testutil.TruncateTable(t, pool, "users")

		userID := testutil.ValidUserID()
		CreateUser(t, pool, userID, "column-move-down@example.com")

		board := testutil.ValidBoard()
		InsertBoard(t, pool, &board)

		first := testutil.ValidColumn(board.ID)
		second := testutil.NewValidColumn(t, board.ID, "In Progress", 2)
		third := testutil.NewValidColumn(t, board.ID, "Done", 3)

		InsertColumn(t, pool, &first)
		InsertColumn(t, pool, &second)
		InsertColumn(t, pool, &third)

		targetPosition, err := domain.NewColumnPosition(3)
		if err != nil {
			t.Fatalf("NewColumnPosition() error = %v", err)
		}

		gotPosition, err := r.Move(context.Background(), board.ID, first.ID, targetPosition)
		if err != nil {
			t.Fatalf("Move() error = %v", err)
		}
		if gotPosition != targetPosition {
			t.Fatalf("Move() position = %v, want %v", gotPosition, targetPosition)
		}

		got := ListColumnsByBoardID(t, pool, board.ID)
		if len(got) != 3 {
			t.Fatalf("got %d columns after move, want 3", len(got))
		}
		if got[0].ID != second.ID || got[0].Position.Int64() != 1 {
			t.Errorf("got id=%s position=%d, want id=%s position=%d", got[0].ID, got[0].Position.Int64(), second.ID, 1)
		}
		if got[1].ID != third.ID || got[1].Position.Int64() != 2 {
			t.Errorf("got id=%s position=%d, want id=%s position=%d", got[1].ID, got[1].Position.Int64(), third.ID, 2)
		}
		if got[2].ID != first.ID || got[2].Position.Int64() != 3 {
			t.Errorf("got id=%s position=%d, want id=%s position=%d", got[2].ID, got[2].Position.Int64(), first.ID, 3)
		}
	})

	t.Run("Success move up", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "columns")
		testutil.TruncateTable(t, pool, "boards")
		testutil.TruncateTable(t, pool, "users")

		userID := testutil.ValidUserID()
		CreateUser(t, pool, userID, "column-move-up@example.com")

		board := testutil.ValidBoard()
		InsertBoard(t, pool, &board)

		first := testutil.ValidColumn(board.ID)
		second := testutil.NewValidColumn(t, board.ID, "In Progress", 2)
		third := testutil.NewValidColumn(t, board.ID, "Done", 3)

		InsertColumn(t, pool, &first)
		InsertColumn(t, pool, &second)
		InsertColumn(t, pool, &third)

		targetPosition, err := domain.NewColumnPosition(1)
		if err != nil {
			t.Fatalf("NewColumnPosition() error = %v", err)
		}

		gotPosition, err := r.Move(context.Background(), board.ID, third.ID, targetPosition)
		if err != nil {
			t.Fatalf("Move() error = %v", err)
		}
		if gotPosition != targetPosition {
			t.Fatalf("Move() position = %v, want %v", gotPosition, targetPosition)
		}

		got := ListColumnsByBoardID(t, pool, board.ID)
		if len(got) != 3 {
			t.Fatalf("got %d columns after move, want 3", len(got))
		}
		if got[0].ID != third.ID || got[0].Position.Int64() != 1 {
			t.Errorf("got id=%s position=%d, want id=%s position=%d", got[0].ID, got[0].Position.Int64(), third.ID, 1)
		}
		if got[1].ID != first.ID || got[1].Position.Int64() != 2 {
			t.Errorf("got id=%s position=%d, want id=%s position=%d", got[1].ID, got[1].Position.Int64(), first.ID, 2)
		}
		if got[2].ID != second.ID || got[2].Position.Int64() != 3 {
			t.Errorf("got id=%s position=%d, want id=%s position=%d", got[2].ID, got[2].Position.Int64(), second.ID, 3)
		}
	})

	t.Run("Success no-op", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "columns")
		testutil.TruncateTable(t, pool, "boards")
		testutil.TruncateTable(t, pool, "users")

		userID := testutil.ValidUserID()
		CreateUser(t, pool, userID, "column-move-noop@example.com")

		board := testutil.ValidBoard()
		InsertBoard(t, pool, &board)

		first := testutil.ValidColumn(board.ID)
		second := testutil.NewValidColumn(t, board.ID, "In Progress", 2)

		InsertColumn(t, pool, &first)
		InsertColumn(t, pool, &second)

		targetPosition, err := domain.NewColumnPosition(2)
		if err != nil {
			t.Fatalf("NewColumnPosition() error = %v", err)
		}

		gotPosition, err := r.Move(context.Background(), board.ID, second.ID, targetPosition)
		if err != nil {
			t.Fatalf("Move() error = %v", err)
		}
		if gotPosition != targetPosition {
			t.Fatalf("Move() position = %v, want %v", gotPosition, targetPosition)
		}

		got := ListColumnsByBoardID(t, pool, board.ID)
		if len(got) != 2 {
			t.Fatalf("got %d columns after no-op move, want 2", len(got))
		}
		if got[0].ID != first.ID || got[0].Position.Int64() != 1 {
			t.Errorf("got id=%s position=%d, want id=%s position=%d", got[0].ID, got[0].Position.Int64(), first.ID, 1)
		}
		if got[1].ID != second.ID || got[1].Position.Int64() != 2 {
			t.Errorf("got id=%s position=%d, want id=%s position=%d", got[1].ID, got[1].Position.Int64(), second.ID, 2)
		}
	})

	t.Run("Index out of bounds", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "columns")
		testutil.TruncateTable(t, pool, "boards")
		testutil.TruncateTable(t, pool, "users")

		userID := testutil.ValidUserID()
		CreateUser(t, pool, userID, "column-move-oob@example.com")

		board := testutil.ValidBoard()
		InsertBoard(t, pool, &board)

		first := testutil.ValidColumn(board.ID)
		second := testutil.NewValidColumn(t, board.ID, "In Progress", 2)
		third := testutil.NewValidColumn(t, board.ID, "Done", 3)

		InsertColumn(t, pool, &first)
		InsertColumn(t, pool, &second)
		InsertColumn(t, pool, &third)

		targetPosition, err := domain.NewColumnPosition(4)
		if err != nil {
			t.Fatalf("NewColumnPosition() error = %v", err)
		}

		_, err = r.Move(context.Background(), board.ID, second.ID, targetPosition)
		if !errors.Is(err, repository.ErrIndexOutOfBounds) {
			t.Fatalf("Move() error = %v, want ErrIndexOutOfBounds", err)
		}

		got := ListColumnsByBoardID(t, pool, board.ID)
		if len(got) != 3 {
			t.Fatalf("got %d columns after failed move, want 3", len(got))
		}
		if got[0].ID != first.ID || got[0].Position.Int64() != 1 {
			t.Errorf("got id=%s position=%d, want id=%s position=%d", got[0].ID, got[0].Position.Int64(), first.ID, 1)
		}
		if got[1].ID != second.ID || got[1].Position.Int64() != 2 {
			t.Errorf("got id=%s position=%d, want id=%s position=%d", got[1].ID, got[1].Position.Int64(), second.ID, 2)
		}
		if got[2].ID != third.ID || got[2].Position.Int64() != 3 {
			t.Errorf("got id=%s position=%d, want id=%s position=%d", got[2].ID, got[2].Position.Int64(), third.ID, 3)
		}
	})

	t.Run("Not found by column id", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "columns")
		testutil.TruncateTable(t, pool, "boards")
		testutil.TruncateTable(t, pool, "users")

		userID := testutil.ValidUserID()
		CreateUser(t, pool, userID, "column-move-missing-col@example.com")

		board := testutil.ValidBoard()
		InsertBoard(t, pool, &board)

		targetPosition, err := domain.NewColumnPosition(1)
		if err != nil {
			t.Fatalf("NewColumnPosition() error = %v", err)
		}

		_, err = r.Move(context.Background(), board.ID, domain.NewColumnID(), targetPosition)
		if !errors.Is(err, repository.ErrRowNotFound) {
			t.Errorf("Move() error = %v, want ErrRowNotFound", err)
		}
	})

	t.Run("Not found by board id", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "columns")
		testutil.TruncateTable(t, pool, "boards")
		testutil.TruncateTable(t, pool, "users")

		userID := testutil.ValidUserID()
		CreateUser(t, pool, userID, "column-move-missing-board@example.com")

		board := testutil.ValidBoard()
		InsertBoard(t, pool, &board)

		created := testutil.ValidColumn(board.ID)
		InsertColumn(t, pool, &created)

		targetPosition, err := domain.NewColumnPosition(1)
		if err != nil {
			t.Fatalf("NewColumnPosition() error = %v", err)
		}

		_, err = r.Move(context.Background(), domain.NewBoardID(), created.ID, targetPosition)
		if !errors.Is(err, repository.ErrRowNotFound) {
			t.Errorf("Move() error = %v, want ErrRowNotFound", err)
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
			t.Fatalf("got %d columns after delete, want 2", len(got))
		}
		if got[0].ID != first.ID {
			t.Errorf("got first column id %q, want %q", got[0].ID, first.ID)
		}
		if got[0].Position.Int64() != 1 {
			t.Errorf("got first position %d, want 1", got[0].Position.Int64())
		}
		if got[1].ID != third.ID {
			t.Errorf("got second column id %q, want %q", got[1].ID, third.ID)
		}
		if got[1].Position.Int64() != 2 {
			t.Errorf("got second position %d after shift, want 2", got[1].Position.Int64())
		}

		_, ok := FindColumnByID(t, pool, second.ID)
		if ok {
			t.Error("got deleted column in DB, want absent")
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
