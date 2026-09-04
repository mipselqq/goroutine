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

	tests := []struct {
		name              string
		useUnrelatedOwner bool
		useMissingOwner   bool
		useUnrelatedBoard bool
		useMissingBoard   bool
		wantErr           error
	}{
		{name: "Success"},
		{name: "Unrelated owner", useUnrelatedOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing owner", useMissingOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Unrelated board", useUnrelatedBoard: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing board", useMissingBoard: true, wantErr: repository.ErrRowNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.TruncateAllTables(t, pool)
			fixture := setupDefaultBoardHierarchy(t, pool)
			boardWithoutColumns := testutil.ValidBoardForOwner(fixture.board.OwnerID)
			CreateBoard(t, pool, &boardWithoutColumns)
			newColumn := testutil.ValidColumn(boardWithoutColumns.ID)

			callerID := fixture.board.OwnerID
			board := boardWithoutColumns
			if tt.useUnrelatedOwner {
				callerID = fixture.unrelatedBoard.OwnerID
			}
			if tt.useMissingOwner {
				callerID = fixture.nonexistentBoard.OwnerID
			}
			if tt.useUnrelatedBoard {
				board = fixture.unrelatedBoard
			}
			if tt.useMissingBoard {
				board = fixture.nonexistentBoard
			}

			column, err := r.Create(context.Background(), callerID, board.ID, newColumn.Name, newColumn.Description)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Create() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if got := ListColumnsByBoardID(t, pool, boardWithoutColumns.ID); len(got) != 0 {
					t.Errorf("got %d columns, want 0", len(got))
				}
				return
			}

			if column.ID.IsNil() {
				t.Error("got empty column id, want generated id")
			}
			if column.BoardID != boardWithoutColumns.ID {
				t.Errorf("got boardID %q, want %q", column.BoardID, boardWithoutColumns.ID)
			}
			if column.Name != newColumn.Name {
				t.Errorf("got name %q, want %q", column.Name, newColumn.Name)
			}
			if column.Description != newColumn.Description {
				t.Errorf("got description %q, want %q", column.Description, newColumn.Description)
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

			storedColumns := ListColumnsByBoardID(t, pool, boardWithoutColumns.ID)
			if len(storedColumns) != 1 {
				t.Fatalf("ListColumnsByBoardID() returned %d columns, want exactly 1", len(storedColumns))
			}
			if diff := cmp.Diff(column, storedColumns[0], testutil.CmpAllowUnexported()); diff != "" {
				t.Errorf("got stored column mismatch (-want +got):\n%s", diff)
			}

			AssertOutboxEvents(t, pool, []WantOutboxEvent{{
				RecipientUserID: boardWithoutColumns.OwnerID,
				Type:            "column.created",
				Payload: struct {
					CallerEmail string `json:"callerEmail"`
					BoardName   string `json:"boardName"`
					ColumnName  string `json:"columnName"`
				}{
					CallerEmail: testutil.ValidEmail().String(),
					BoardName:   boardWithoutColumns.Name.String(),
					ColumnName:  column.Name.String(),
				},
			}})
		})
	}
}

func TestColumnRepository_Create_AppendsPosition(t *testing.T) {
	pool, r := columnRepoPrelude(t)
	testutil.TruncateAllTables(t, pool)

	fixture := setupDefaultBoardHierarchy(t, pool)

	newColumn := testutil.NewValidColumn(t, fixture.board.ID, "Done", 3)

	createdColumn, err := r.Create(
		context.Background(),
		fixture.board.OwnerID,
		fixture.board.ID,
		newColumn.Name,
		newColumn.Description,
	)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if createdColumn.Position.Int64() != 3 {
		t.Errorf("got position %d, want 3", createdColumn.Position.Int64())
	}
}

func TestColumnRepository_ListByBoardID(t *testing.T) {
	pool, r := columnRepoPrelude(t)

	tests := []struct {
		name              string
		useUnrelatedOwner bool
		useMissingOwner   bool
		useUnrelatedBoard bool
		useMissingBoard   bool
		wantErr           error
	}{
		{name: "Success empty"},
		{name: "Unrelated owner", useUnrelatedOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing owner", useMissingOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Unrelated board", useUnrelatedBoard: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing board", useMissingBoard: true, wantErr: repository.ErrRowNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.TruncateAllTables(t, pool)
			fixture := setupDefaultBoardHierarchy(t, pool)
			boardWithoutColumns := testutil.ValidBoardForOwner(fixture.board.OwnerID)
			CreateBoard(t, pool, &boardWithoutColumns)

			callerID := fixture.board.OwnerID
			board := boardWithoutColumns
			if tt.useUnrelatedOwner {
				callerID = fixture.unrelatedBoard.OwnerID
			}
			if tt.useMissingOwner {
				callerID = fixture.nonexistentBoard.OwnerID
			}
			if tt.useUnrelatedBoard {
				board = fixture.unrelatedBoard
			}
			if tt.useMissingBoard {
				board = fixture.nonexistentBoard
			}

			columns, err := r.ListByBoardID(context.Background(), callerID, board.ID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ListByBoardID() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && len(columns) != 0 {
				t.Errorf("got %d columns, want 0", len(columns))
			}
		})
	}

	t.Run("Success ordered and filtered by board", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)
		fixture := setupDefaultBoardHierarchy(t, pool)

		got, err := r.ListByBoardID(context.Background(), fixture.board.OwnerID, fixture.board.ID)
		if err != nil {
			t.Fatalf("ListByBoardID() error = %v", err)
		}

		want := []domain.Column{fixture.column, fixture.siblingColumn}
		if diff := cmp.Diff(want, got, testutil.CmpAllowUnexported()); diff != "" {
			t.Errorf("ListByBoardID() mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestColumnRepository_Get(t *testing.T) {
	pool, r := columnRepoPrelude(t)

	tests := []struct {
		name               string
		useUnrelatedOwner  bool
		useMissingOwner    bool
		useUnrelatedBoard  bool
		useMissingBoard    bool
		useUnrelatedColumn bool
		useMissingColumn   bool
		wantErr            error
	}{
		{name: "Success"},
		{name: "Unrelated owner", useUnrelatedOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing owner", useMissingOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Unrelated board", useUnrelatedBoard: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing board", useMissingBoard: true, wantErr: repository.ErrRowNotFound},
		{name: "Unrelated column", useUnrelatedColumn: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing column", useMissingColumn: true, wantErr: repository.ErrRowNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.TruncateAllTables(t, pool)
			fixture := setupDefaultBoardHierarchy(t, pool)

			callerID := fixture.board.OwnerID
			board := fixture.board
			column := fixture.column
			if tt.useUnrelatedOwner {
				callerID = fixture.unrelatedBoard.OwnerID
			}
			if tt.useMissingOwner {
				callerID = fixture.nonexistentBoard.OwnerID
			}
			if tt.useUnrelatedBoard {
				board = fixture.unrelatedBoard
			}
			if tt.useMissingBoard {
				board = fixture.nonexistentBoard
			}
			if tt.useUnrelatedColumn {
				column = fixture.unrelatedColumn
			}
			if tt.useMissingColumn {
				column = fixture.nonexistentColumn
			}

			got, err := r.Get(context.Background(), callerID, board.ID, column.ID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Get() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if diff := cmp.Diff(fixture.column, got, testutil.CmpAllowUnexported()); diff != "" {
					t.Errorf("Get() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
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
		if len(storedColumns) != 2 {
			t.Fatalf("ListColumnsByBoardID() returned %d columns, want exactly 2", len(storedColumns))
		}
		for _, storedColumn := range storedColumns {
			if storedColumn.ID == got.ID {
				if diff := cmp.Diff(got, storedColumn, testutil.CmpAllowUnexported()); diff != "" {
					t.Errorf("got stored column mismatch (-want +got):\n%s", diff)
				}
				return
			}
		}
		t.Errorf("updated column %v not found", got.ID)
	}

	successUpdateOutboxEvents := func(ownerID domain.UserID, columnName domain.ColumnName) []WantOutboxEvent {
		return []WantOutboxEvent{{
			RecipientUserID: ownerID,
			Type:            "column.updated",
			Payload: struct {
				CallerEmail string `json:"callerEmail"`
				BoardName   string `json:"boardName"`
				ColumnName  string `json:"columnName"`
			}{
				CallerEmail: testutil.ValidEmail().String(),
				BoardName:   testutil.ValidBoardName().String(),
				ColumnName:  columnName.String(),
			},
		}}
	}

	tests := []struct {
		name               string
		useUnrelatedOwner  bool
		useMissingOwner    bool
		useUnrelatedBoard  bool
		useMissingBoard    bool
		useUnrelatedColumn bool
		useMissingColumn   bool
		wantErr            error
	}{
		{name: "Success"},
		{name: "Unrelated owner", useUnrelatedOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing owner", useMissingOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Unrelated board", useUnrelatedBoard: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing board", useMissingBoard: true, wantErr: repository.ErrRowNotFound},
		{name: "Unrelated column", useUnrelatedColumn: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing column", useMissingColumn: true, wantErr: repository.ErrRowNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.TruncateAllTables(t, pool)
			fixture := setupDefaultBoardHierarchy(t, pool)
			want := testutil.UpdateValidColumn(t, &fixture.column, "Renamed", fixture.column.Description.String(), fixture.column.UpdatedAt)

			callerID := fixture.board.OwnerID
			board := fixture.board
			column := fixture.column
			if tt.useUnrelatedOwner {
				callerID = fixture.unrelatedBoard.OwnerID
			}
			if tt.useMissingOwner {
				callerID = fixture.nonexistentBoard.OwnerID
			}
			if tt.useUnrelatedBoard {
				board = fixture.unrelatedBoard
			}
			if tt.useMissingBoard {
				board = fixture.nonexistentBoard
			}
			if tt.useUnrelatedColumn {
				column = fixture.unrelatedColumn
			}
			if tt.useMissingColumn {
				column = fixture.nonexistentColumn
			}

			got, err := r.Update(context.Background(), callerID, board.ID, column.ID, &want.Name, nil)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Update() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				assertUpdatedColumn(t, got, want)
				AssertOutboxEvents(t, pool, successUpdateOutboxEvents(fixture.board.OwnerID, got.Name))
				return
			}

			if diff := cmp.Diff([]domain.Column{fixture.column, fixture.siblingColumn}, ListColumnsByBoardID(t, pool, fixture.board.ID), testutil.CmpAllowUnexported()); diff != "" {
				t.Errorf("stored columns mismatch (-want +got):\n%s", diff)
			}
		})
	}

	t.Run("Success description only", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		fixture := setupDefaultBoardHierarchy(t, pool)
		column := fixture.column

		newDesc, err := domain.NewColumnDescription("Updated column body")
		if err != nil {
			t.Fatalf("NewColumnDescription() error = %v", err)
		}
		updated, err := r.Update(context.Background(), fixture.board.OwnerID, fixture.board.ID, column.ID, nil, &newDesc)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		if updated.Name != column.Name {
			t.Errorf("got name %q, want %q", updated.Name, column.Name)
		}
		if updated.Description != newDesc {
			t.Errorf("got description %q, want %q", updated.Description, newDesc)
		}
		AssertOutboxEvents(t, pool, successUpdateOutboxEvents(fixture.board.OwnerID, updated.Name))
		storedColumns := ListColumnsByBoardID(t, pool, column.BoardID)
		if len(storedColumns) != 2 {
			t.Fatalf("ListColumnsByBoardID() returned %d columns, want exactly 2", len(storedColumns))
		}
		for _, storedColumn := range storedColumns {
			if storedColumn.ID == column.ID {
				if storedColumn.Description != newDesc {
					t.Errorf("stored description %q, want %q", storedColumn.Description, newDesc)
				}
				return
			}
		}
		t.Errorf("updated column %v not found", column.ID)
	})

	t.Run("Success no changes", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		fixture := setupDefaultBoardHierarchy(t, pool)

		got, err := r.Update(context.Background(), fixture.board.OwnerID, fixture.board.ID, fixture.column.ID, nil, nil)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if diff := cmp.Diff(fixture.column, got, testutil.CmpAllowUnexported()); diff != "" {
			t.Errorf("Update() mismatch (-want +got):\n%s", diff)
		}
		AssertOutboxEvents(t, pool, []WantOutboxEvent{})
	})
}

func TestColumnRepository_Move(t *testing.T) {
	pool, r := columnRepoPrelude(t)

	tests := []struct {
		name               string
		useUnrelatedOwner  bool
		useMissingOwner    bool
		useUnrelatedBoard  bool
		useMissingBoard    bool
		useUnrelatedColumn bool
		useMissingColumn   bool
		wantErr            error
	}{
		{name: "Success move down"},
		{name: "Unrelated owner", useUnrelatedOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing owner", useMissingOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Unrelated board", useUnrelatedBoard: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing board", useMissingBoard: true, wantErr: repository.ErrRowNotFound},
		{name: "Unrelated column", useUnrelatedColumn: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing column", useMissingColumn: true, wantErr: repository.ErrRowNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.TruncateAllTables(t, pool)
			fixture := setupDefaultBoardHierarchy(t, pool)
			firstColumn := fixture.column
			secondColumn := fixture.siblingColumn
			thirdColumn := testutil.NewValidColumn(t, fixture.board.ID, "Done", 3)
			CreateColumn(t, pool, &thirdColumn)

			callerID := fixture.board.OwnerID
			board := fixture.board
			column := firstColumn
			if tt.useUnrelatedOwner {
				callerID = fixture.unrelatedBoard.OwnerID
			}
			if tt.useMissingOwner {
				callerID = fixture.nonexistentBoard.OwnerID
			}
			if tt.useUnrelatedBoard {
				board = fixture.unrelatedBoard
			}
			if tt.useMissingBoard {
				board = fixture.nonexistentBoard
			}
			if tt.useUnrelatedColumn {
				column = fixture.unrelatedColumn
			}
			if tt.useMissingColumn {
				column = fixture.nonexistentColumn
			}

			destinationPosition := testutil.NewValidColumnPosition(t, 3)
			gotPosition, err := r.Move(context.Background(), callerID, board.ID, column.ID, destinationPosition)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Move() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && gotPosition != destinationPosition {
				t.Errorf("Move() position = %v, want %v", gotPosition, destinationPosition)
			}

			got := ListColumnsByBoardID(t, pool, fixture.board.ID)
			if len(got) != 3 {
				t.Fatalf("got %d columns after move, want 3", len(got))
			}
			if tt.wantErr == nil {
				assertColumnIDAndPosition(t, &got[0], secondColumn.ID, 1)
				assertColumnIDAndPosition(t, &got[1], thirdColumn.ID, 2)
				assertColumnIDAndPosition(t, &got[2], firstColumn.ID, 3)

				AssertOutboxEvents(t, pool, []WantOutboxEvent{{
					RecipientUserID: fixture.board.OwnerID,
					Type:            "column.moved",
					Payload: struct {
						CallerEmail    string `json:"callerEmail"`
						BoardName      string `json:"boardName"`
						ColumnName     string `json:"columnName"`
						SourcePosition int64  `json:"sourcePosition"`
						TargetPosition int64  `json:"targetPosition"`
					}{
						CallerEmail:    testutil.ValidEmail().String(),
						BoardName:      fixture.board.Name.String(),
						ColumnName:     firstColumn.Name.String(),
						SourcePosition: firstColumn.Position.Int64(),
						TargetPosition: destinationPosition.Int64(),
					},
				}})
				return
			}
			assertColumnIDAndPosition(t, &got[0], firstColumn.ID, 1)
			assertColumnIDAndPosition(t, &got[1], secondColumn.ID, 2)
			assertColumnIDAndPosition(t, &got[2], thirdColumn.ID, 3)
		})
	}

	t.Run("Success move up", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		fixture := setupDefaultBoardHierarchy(t, pool)
		board := fixture.board
		secondColumn := fixture.siblingColumn
		thirdColumn := testutil.NewValidColumn(t, board.ID, "Done", 3)

		CreateColumn(t, pool, &thirdColumn)

		destinationPosition := testutil.NewValidColumnPosition(t, 1)

		gotPosition, err := r.Move(context.Background(), board.OwnerID, board.ID, thirdColumn.ID, destinationPosition)
		if err != nil {
			t.Fatalf("Move() error = %v", err)
		}
		if gotPosition != destinationPosition {
			t.Fatalf("Move() position = %v, want %v", gotPosition, destinationPosition)
		}

		got := ListColumnsByBoardID(t, pool, board.ID)
		if len(got) != 3 {
			t.Fatalf("got %d columns after move, want 3", len(got))
		}
		assertColumnIDAndPosition(t, &got[0], thirdColumn.ID, 1)
		assertColumnIDAndPosition(t, &got[1], fixture.column.ID, 2)
		assertColumnIDAndPosition(t, &got[2], secondColumn.ID, 3)
	})

	t.Run("Success no-op", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		fixture := setupDefaultBoardHierarchy(t, pool)
		board := fixture.board
		secondColumn := fixture.siblingColumn

		destinationPosition := testutil.NewValidColumnPosition(t, 2)

		gotPosition, err := r.Move(context.Background(), board.OwnerID, board.ID, secondColumn.ID, destinationPosition)
		if err != nil {
			t.Fatalf("Move() error = %v", err)
		}
		if gotPosition != destinationPosition {
			t.Fatalf("Move() position = %v, want %v", gotPosition, destinationPosition)
		}

		got := ListColumnsByBoardID(t, pool, board.ID)
		if len(got) != 2 {
			t.Fatalf("got %d columns after no-op move, want 2", len(got))
		}
		assertColumnIDAndPosition(t, &got[0], fixture.column.ID, 1)
		assertColumnIDAndPosition(t, &got[1], secondColumn.ID, 2)
		AssertOutboxEvents(t, pool, []WantOutboxEvent{})
	})

	t.Run("Index out of bounds", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		fixture := setupDefaultBoardHierarchy(t, pool)
		board := fixture.board
		secondColumn := fixture.siblingColumn
		thirdColumn := testutil.NewValidColumn(t, board.ID, "Done", 3)

		CreateColumn(t, pool, &thirdColumn)

		destinationPosition := testutil.NewValidColumnPosition(t, 4)

		_, err := r.Move(context.Background(), board.OwnerID, board.ID, secondColumn.ID, destinationPosition)
		if !errors.Is(err, repository.ErrIndexOutOfBounds) {
			t.Fatalf("Move() error = %v, want ErrIndexOutOfBounds", err)
		}

		got := ListColumnsByBoardID(t, pool, board.ID)
		if len(got) != 3 {
			t.Fatalf("got %d columns after failed move, want 3", len(got))
		}
		assertColumnIDAndPosition(t, &got[0], fixture.column.ID, 1)
		assertColumnIDAndPosition(t, &got[1], secondColumn.ID, 2)
		assertColumnIDAndPosition(t, &got[2], thirdColumn.ID, 3)
	})
}

func TestColumnRepository_Delete(t *testing.T) {
	pool, r := columnRepoPrelude(t)

	tests := []struct {
		name               string
		useUnrelatedOwner  bool
		useMissingOwner    bool
		useUnrelatedBoard  bool
		useMissingBoard    bool
		useUnrelatedColumn bool
		useMissingColumn   bool
		wantErr            error
	}{
		{name: "Success shift positions"},
		{name: "Unrelated owner", useUnrelatedOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing owner", useMissingOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Unrelated board", useUnrelatedBoard: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing board", useMissingBoard: true, wantErr: repository.ErrRowNotFound},
		{name: "Unrelated column", useUnrelatedColumn: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing column", useMissingColumn: true, wantErr: repository.ErrRowNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.TruncateAllTables(t, pool)
			fixture := setupDefaultBoardHierarchy(t, pool)
			firstColumn := fixture.column
			secondColumn := fixture.siblingColumn
			thirdColumn := testutil.NewValidColumn(t, fixture.board.ID, "Done", 3)
			CreateColumn(t, pool, &thirdColumn)

			callerID := fixture.board.OwnerID
			board := fixture.board
			column := secondColumn
			if tt.useUnrelatedOwner {
				callerID = fixture.unrelatedBoard.OwnerID
			}
			if tt.useMissingOwner {
				callerID = fixture.nonexistentBoard.OwnerID
			}
			if tt.useUnrelatedBoard {
				board = fixture.unrelatedBoard
			}
			if tt.useMissingBoard {
				board = fixture.nonexistentBoard
			}
			if tt.useUnrelatedColumn {
				column = fixture.unrelatedColumn
			}
			if tt.useMissingColumn {
				column = fixture.nonexistentColumn
			}

			err := r.Delete(context.Background(), callerID, board.ID, column.ID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Delete() error = %v, want %v", err, tt.wantErr)
			}

			got := ListColumnsByBoardID(t, pool, fixture.board.ID)
			if tt.wantErr == nil {
				if len(got) != 2 {
					t.Fatalf("got %d columns after delete, want 2", len(got))
				}
				assertColumnIDAndPosition(t, &got[0], firstColumn.ID, 1)
				assertColumnIDAndPosition(t, &got[1], thirdColumn.ID, 2)

				AssertOutboxEvents(t, pool, []WantOutboxEvent{{
					RecipientUserID: fixture.board.OwnerID,
					Type:            "column.deleted",
					Payload: struct {
						CallerEmail string `json:"callerEmail"`
						BoardName   string `json:"boardName"`
						ColumnName  string `json:"columnName"`
					}{
						CallerEmail: testutil.ValidEmail().String(),
						BoardName:   fixture.board.Name.String(),
						ColumnName:  secondColumn.Name.String(),
					},
				}})
				return
			}
			if len(got) != 3 {
				t.Fatalf("got %d columns after failed delete, want 3", len(got))
			}
			assertColumnIDAndPosition(t, &got[0], firstColumn.ID, 1)
			assertColumnIDAndPosition(t, &got[1], secondColumn.ID, 2)
			assertColumnIDAndPosition(t, &got[2], thirdColumn.ID, 3)
		})
	}
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

	pool := testutil.SetupPostgres(t)
	t.Cleanup(func() { pool.Close() })

	return pool, repository.NewPGColumn(pool)
}
