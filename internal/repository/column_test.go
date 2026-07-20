//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5/pgxpool"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/testutil"
)

func TestColumnRepository_Create(t *testing.T) {
	pool, r := columnRepoPrelude(t)

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board := insertFixedUserAndBoard(t, pool)

		validColumn := testutil.ValidColumn(board.ID)

		column, err := r.Create(
			context.Background(),
			board.ID,
			validColumn.Name,
			validColumn.Description,
		)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if column.ID.IsNil() {
			t.Error("got empty column id, want generated id")
		}
		if column.BoardID != board.ID {
			t.Errorf("got boardID %q, want %q", column.BoardID, board.ID)
		}
		if column.Name != validColumn.Name {
			t.Errorf("got name %q, want %q", column.Name, validColumn.Name)
		}
		if column.Description != validColumn.Description {
			t.Errorf("got description %q, want %q", column.Description, validColumn.Description)
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

		storedColumns := ListColumnsByBoardID(t, pool, board.ID)
		if len(storedColumns) != 1 {
			t.Fatalf("ListColumnsByBoardID() returned %d columns, want exactly 1", len(storedColumns))
		}
		if diff := cmp.Diff(column, storedColumns[0], testutil.CmpAllowUnexported()); diff != "" {
			t.Errorf("got stored column mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestColumnRepository_Create_AppendsPosition(t *testing.T) {
	pool, r := columnRepoPrelude(t)

	testutil.TruncateAllTables(t, pool)

	board := insertFixedUserAndBoard(t, pool)

	existing := testutil.ValidColumn(board.ID)
	CreateColumn(t, pool, &existing)

	toCreate := testutil.ValidColumn(board.ID)
	toCreate = testutil.UpdateValidColumn(t, &toCreate, "Done", toCreate.Description.String(), toCreate.UpdatedAt)

	second, err := r.Create(
		context.Background(),
		board.ID,
		toCreate.Name,
		toCreate.Description,
	)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if second.Position.Int64() != 2 {
		t.Errorf("got second position %d, want 2", second.Position.Int64())
	}
}

func TestColumnRepository_ListByBoardID(t *testing.T) {
	pool, r := columnRepoPrelude(t)

	t.Run("Success empty", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board := insertFixedUserAndBoard(t, pool)

		columns, err := r.ListByBoardID(context.Background(), board.ID)
		if err != nil {
			t.Fatalf("ListByBoardID() error = %v", err)
		}
		if len(columns) != 0 {
			t.Fatalf("got %d columns, want 0", len(columns))
		}
	})

	t.Run("Success ordered and filtered by board", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		CreateFixedUser(t, pool)

		boardA := testutil.ValidBoard()
		CreateBoard(t, pool, &boardA)

		boardB := testutil.ValidBoard()
		CreateBoard(t, pool, &boardB)

		first := testutil.ValidColumn(boardA.ID)
		second := testutil.NewValidColumn(t, boardA.ID, "In Progress", 2)
		otherBoardColumn := testutil.NewValidColumn(t, boardB.ID, "Done", 1)

		CreateColumn(t, pool, &second)
		CreateColumn(t, pool, &first)
		CreateColumn(t, pool, &otherBoardColumn)

		got, err := r.ListByBoardID(context.Background(), boardA.ID)
		if err != nil {
			t.Fatalf("ListByBoardID() error = %v", err)
		}

		want := []domain.Column{first, second}
		if diff := cmp.Diff(want, got, testutil.CmpAllowUnexported()); diff != "" {
			t.Errorf("ListByBoardID() mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestColumnRepository_Get(t *testing.T) {
	pool, r := columnRepoPrelude(t)

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board := insertFixedUserAndBoard(t, pool)

		created := testutil.ValidColumn(board.ID)
		CreateColumn(t, pool, &created)

		got, err := r.Get(context.Background(), created.ID)
		if err != nil {
			t.Fatalf("Get() error = %v", err)
		}
		if diff := cmp.Diff(created, got, testutil.CmpAllowUnexported()); diff != "" {
			t.Errorf("Get() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("Not found", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		_, err := r.Get(context.Background(), domain.NewColumnID())
		assertErrRowNotFound(t, err)
	})
}

func TestColumnRepository_Update(t *testing.T) {
	pool, r := columnRepoPrelude(t)

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
		if got.Description != want.Description {
			t.Errorf("got description %q, want %q", got.Description, want.Description)
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

		storedColumns := ListColumnsByBoardID(t, pool, want.BoardID)
		if len(storedColumns) != 1 {
			t.Fatalf("ListColumnsByBoardID() returned %d columns, want exactly 1", len(storedColumns))
		}
		if diff := cmp.Diff(got, storedColumns[0], testutil.CmpAllowUnexported()); diff != "" {
			t.Errorf("got stored column mismatch (-want +got):\n%s", diff)
		}
	}

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board := insertFixedUserAndBoard(t, pool)

		created := testutil.ValidColumn(board.ID)
		createdAtBeforeUpdate := time.Now().UTC()
		updatedAtBeforeUpdate := createdAtBeforeUpdate
		created.CreatedAt = createdAtBeforeUpdate
		created.UpdatedAt = updatedAtBeforeUpdate
		CreateColumn(t, pool, &created)

		want := testutil.UpdateValidColumn(t, &created, "Renamed", created.Description.String(), testutil.Fixed5mFromNow())
		updated, err := r.Update(context.Background(), board.ID, created.ID, &want.Name, nil)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		assertUpdatedColumn(t, updated, want)
	})

	t.Run("Success description only", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board := insertFixedUserAndBoard(t, pool)

		created := testutil.ValidColumn(board.ID)
		createdAtBeforeUpdate := time.Now().UTC()
		created.CreatedAt = createdAtBeforeUpdate
		created.UpdatedAt = createdAtBeforeUpdate
		CreateColumn(t, pool, &created)

		newDesc, err := domain.NewColumnDescription("Updated column body")
		if err != nil {
			t.Fatalf("NewColumnDescription() error = %v", err)
		}
		updated, err := r.Update(context.Background(), board.ID, created.ID, nil, &newDesc)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if updated.Name != created.Name {
			t.Errorf("got name %q, want %q", updated.Name, created.Name)
		}
		if updated.Description != newDesc {
			t.Errorf("got description %q, want %q", updated.Description, newDesc)
		}
		storedColumns := ListColumnsByBoardID(t, pool, created.BoardID)
		if len(storedColumns) != 1 {
			t.Fatalf("ListColumnsByBoardID() returned %d columns, want exactly 1", len(storedColumns))
		}
		storedColumn := storedColumns[0]
		if storedColumn.Description != newDesc {
			t.Errorf("stored description %q, want %q", storedColumn.Description, newDesc)
		}
	})

	t.Run("Not found by column id", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board := insertFixedUserAndBoard(t, pool)

		updatedName, _ := domain.NewColumnName("Renamed")
		_, err := r.Update(context.Background(), board.ID, domain.NewColumnID(), &updatedName, nil)
		assertErrRowNotFound(t, err)
	})

	t.Run("Not found by board id", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board := insertFixedUserAndBoard(t, pool)

		created := testutil.ValidColumn(board.ID)
		CreateColumn(t, pool, &created)

		want := testutil.UpdateValidColumn(t, &created, "Renamed", created.Description.String(), testutil.Fixed5mFromNow())
		_, err := r.Update(context.Background(), domain.NewBoardID(), created.ID, &want.Name, nil)
		assertErrRowNotFound(t, err)
	})
}

func TestColumnRepository_Move(t *testing.T) {
	pool, r := columnRepoPrelude(t)

	t.Run("Success move down", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board := insertFixedUserAndBoard(t, pool)

		first := testutil.ValidColumn(board.ID)
		second := testutil.NewValidColumn(t, board.ID, "In Progress", 2)
		third := testutil.NewValidColumn(t, board.ID, "Done", 3)

		CreateColumn(t, pool, &third)
		CreateColumn(t, pool, &first)
		CreateColumn(t, pool, &second)

		targetPosition := testutil.NewValidColumnPosition(t, 3)

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
		assertColumnIDAndPosition(t, &got[0], second.ID, 1)
		assertColumnIDAndPosition(t, &got[1], third.ID, 2)
		assertColumnIDAndPosition(t, &got[2], first.ID, 3)
	})

	t.Run("Success move up", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board := insertFixedUserAndBoard(t, pool)

		first := testutil.ValidColumn(board.ID)
		second := testutil.NewValidColumn(t, board.ID, "In Progress", 2)
		third := testutil.NewValidColumn(t, board.ID, "Done", 3)

		CreateColumn(t, pool, &second)
		CreateColumn(t, pool, &third)
		CreateColumn(t, pool, &first)

		targetPosition := testutil.NewValidColumnPosition(t, 1)

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
		assertColumnIDAndPosition(t, &got[0], third.ID, 1)
		assertColumnIDAndPosition(t, &got[1], first.ID, 2)
		assertColumnIDAndPosition(t, &got[2], second.ID, 3)
	})

	t.Run("Success no-op", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board := insertFixedUserAndBoard(t, pool)

		first := testutil.ValidColumn(board.ID)
		second := testutil.NewValidColumn(t, board.ID, "In Progress", 2)

		CreateColumn(t, pool, &second)
		CreateColumn(t, pool, &first)

		targetPosition := testutil.NewValidColumnPosition(t, 2)

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
		assertColumnIDAndPosition(t, &got[0], first.ID, 1)
		assertColumnIDAndPosition(t, &got[1], second.ID, 2)
	})

	t.Run("Index out of bounds", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board := insertFixedUserAndBoard(t, pool)

		first := testutil.ValidColumn(board.ID)
		second := testutil.NewValidColumn(t, board.ID, "In Progress", 2)
		third := testutil.NewValidColumn(t, board.ID, "Done", 3)

		CreateColumn(t, pool, &second)
		CreateColumn(t, pool, &third)
		CreateColumn(t, pool, &first)

		targetPosition := testutil.NewValidColumnPosition(t, 4)

		_, err := r.Move(context.Background(), board.ID, second.ID, targetPosition)
		if !errors.Is(err, repository.ErrIndexOutOfBounds) {
			t.Fatalf("Move() error = %v, want ErrIndexOutOfBounds", err)
		}

		got := ListColumnsByBoardID(t, pool, board.ID)
		if len(got) != 3 {
			t.Fatalf("got %d columns after failed move, want 3", len(got))
		}
		assertColumnIDAndPosition(t, &got[0], first.ID, 1)
		assertColumnIDAndPosition(t, &got[1], second.ID, 2)
		assertColumnIDAndPosition(t, &got[2], third.ID, 3)
	})

	t.Run("Not found by column id", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board := insertFixedUserAndBoard(t, pool)

		targetPosition := testutil.NewValidColumnPosition(t, 1)

		_, err := r.Move(context.Background(), board.ID, domain.NewColumnID(), targetPosition)
		assertErrRowNotFound(t, err)
	})

	t.Run("Not found by board id", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board := insertFixedUserAndBoard(t, pool)

		created := testutil.ValidColumn(board.ID)
		CreateColumn(t, pool, &created)

		targetPosition := testutil.NewValidColumnPosition(t, 1)

		_, err := r.Move(context.Background(), domain.NewBoardID(), created.ID, targetPosition)
		assertErrRowNotFound(t, err)
	})
}

func TestColumnRepository_Delete(t *testing.T) {
	pool, r := columnRepoPrelude(t)

	t.Run("Success shift positions", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board := insertFixedUserAndBoard(t, pool)

		first := testutil.ValidColumn(board.ID)
		second := testutil.NewValidColumn(t, board.ID, "In Progress", 2)
		third := testutil.NewValidColumn(t, board.ID, "Done", 3)

		CreateColumn(t, pool, &third)
		CreateColumn(t, pool, &first)
		CreateColumn(t, pool, &second)

		err := r.Delete(context.Background(), board.ID, second.ID)
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		got := ListColumnsByBoardID(t, pool, board.ID)

		if len(got) != 2 {
			t.Fatalf("got %d columns after delete, want 2", len(got))
		}
		assertColumnIDAndPosition(t, &got[0], first.ID, 1)
		assertColumnIDAndPosition(t, &got[1], third.ID, 2)
	})

	t.Run("Not found by column id", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board := insertFixedUserAndBoard(t, pool)

		err := r.Delete(context.Background(), board.ID, domain.NewColumnID())
		assertErrRowNotFound(t, err)
	})

	t.Run("Not found by board id", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board := insertFixedUserAndBoard(t, pool)

		created := testutil.ValidColumn(board.ID)
		CreateColumn(t, pool, &created)

		err := r.Delete(context.Background(), domain.NewBoardID(), created.ID)
		assertErrRowNotFound(t, err)
	})
}

func assertColumnIDAndPosition(t *testing.T, col *domain.Column, wantID domain.ColumnID, wantPos int64) {
	t.Helper()

	if col.ID != wantID {
		t.Errorf("got id %q, want %q", col.ID, wantID)
	}
	if col.Position.Int64() != wantPos {
		t.Errorf("got position %d, want %d", col.Position.Int64(), wantPos)
	}
}

func columnRepoPrelude(t *testing.T) (*pgxpool.Pool, *repository.PGColumn) {
	t.Helper()

	pool := testutil.SetupPostgres(t, "../../migrations")
	t.Cleanup(func() { pool.Close() })

	return pool, repository.NewPGColumn(pool)
}
