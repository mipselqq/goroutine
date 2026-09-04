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

func TestBoardRepository_Create(t *testing.T) {
	pool, r := boardRepoPrelude(t)

	boardName := testutil.ValidBoardName()
	boardDescription := testutil.ValidBoardDescription()

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)
		fixture := setupDefaultBoardHierarchy(t, pool)

		board, err := r.Create(context.Background(), fixture.board.OwnerID, boardName, boardDescription)
		if err != nil {
			t.Errorf("Create() error = %v", err)
		}
		if board.ID.IsNil() {
			t.Errorf("got empty board ID, want generated ID")
		}
		if board.OwnerID != fixture.board.OwnerID {
			t.Errorf("got owner ID %q, want %q", board.OwnerID, fixture.board.OwnerID)
		}
		if board.Name != boardName {
			t.Errorf("got name %q, want %q", board.Name, boardName)
		}
		if board.Description != boardDescription {
			t.Errorf("got description %q, want %q", board.Description, boardDescription)
		}
		if board.CreatedAt.IsZero() {
			t.Errorf("got zero createdAt, want set value")
		}
		if board.UpdatedAt.IsZero() {
			t.Errorf("got zero updatedAt, want set value")
		}
		if !board.CreatedAt.Equal(board.UpdatedAt) {
			t.Errorf("got createdAt=%v updatedAt=%v, want equal", board.CreatedAt, board.UpdatedAt)
		}
		AssertTimestampPrecisionAtLeastMillis(t, pool, "boards", "created_at", "updated_at")

		for _, storedBoard := range ListBoards(t, pool) {
			if storedBoard.ID == board.ID {
				if diff := cmp.Diff(board, storedBoard, testutil.CmpAllowUnexported()); diff != "" {
					t.Errorf("stored board mismatch (-returned +stored):\n%s", diff)
				}

				AssertOutboxEvents(t, pool, []WantOutboxEvent{{
					RecipientUserID: fixture.board.OwnerID,
					Type:            "board.created",
					Payload: struct {
						CallerEmail string `json:"callerEmail"`
						BoardName   string `json:"boardName"`
					}{
						CallerEmail: testutil.ValidEmail().String(),
						BoardName:   board.Name.String(),
					},
				}})
				return
			}
		}
		t.Errorf("created board %v not found", board.ID)
	})
}

func TestBoardRepository_Get(t *testing.T) {
	pool, r := boardRepoPrelude(t)

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

			callerID := fixture.board.OwnerID
			board := fixture.board
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

			got, err := r.Get(context.Background(), callerID, board.ID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Get() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if diff := cmp.Diff(fixture.board, got, testutil.CmpAllowUnexported()); diff != "" {
					t.Errorf("Get() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestBoardRepository_ListByOwnerID(t *testing.T) {
	pool, r := boardRepoPrelude(t)

	tests := []struct {
		name             string
		useUnrelatedUser bool
		useMissingUser   bool
	}{
		{name: "Success"},
		{name: "Unrelated user", useUnrelatedUser: true},
		{name: "Missing user", useMissingUser: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.TruncateAllTables(t, pool)
			fixture := setupDefaultBoardHierarchy(t, pool)

			ownerID := fixture.board.OwnerID
			want := []domain.Board{fixture.board}
			if tt.useUnrelatedUser {
				ownerID = fixture.unrelatedBoard.OwnerID
				want = []domain.Board{fixture.unrelatedBoard}
			}
			if tt.useMissingUser {
				ownerID = fixture.nonexistentBoard.OwnerID
				want = nil
			}

			got, err := r.ListByOwnerID(context.Background(), ownerID)
			if err != nil {
				t.Errorf("ListByOwnerID() error = %v", err)
			}
			if diff := cmp.Diff(want, got, testutil.CmpAllowUnexported()); diff != "" {
				t.Errorf("ListByOwnerID() mismatch (-want +got):\n%s", diff)
			}
		})
	}

	t.Run("Success ordered and filtered by owner", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)
		fixture := setupDefaultBoardHierarchy(t, pool)

		newerBoard := testutil.ValidBoardForOwner(fixture.board.OwnerID)
		newerBoard.CreatedAt = testutil.Fixed5mFromNow()
		newerBoard.UpdatedAt = testutil.Fixed5mFromNow()
		CreateBoard(t, pool, &newerBoard)

		got, err := r.ListByOwnerID(context.Background(), fixture.board.OwnerID)
		if err != nil {
			t.Fatalf("ListByOwnerID() error = %v", err)
		}

		want := []domain.Board{fixture.board, newerBoard}
		if diff := cmp.Diff(want, got, testutil.CmpAllowUnexported()); diff != "" {
			t.Errorf("ListByOwnerID() mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestBoardRepository_Update(t *testing.T) {
	pool, r := boardRepoPrelude(t)

	assertUpdatedBoard := func(t *testing.T, got domain.Board, want domain.Board) {
		t.Helper()

		if got.ID != want.ID {
			t.Errorf("got id %q, want %q", got.ID, want.ID)
		}
		if got.OwnerID != want.OwnerID {
			t.Errorf("got ownerID %q, want %q", got.OwnerID, want.OwnerID)
		}
		if got.Name != want.Name {
			t.Errorf("got name %q, want %q", got.Name, want.Name)
		}
		if got.Description != want.Description {
			t.Errorf("got description %q, want %q", got.Description, want.Description)
		}
		if !got.CreatedAt.Truncate(time.Millisecond).Equal(want.CreatedAt.Truncate(time.Millisecond)) {
			t.Errorf("got createdAt %v, want %v (at millisecond precision)", got.CreatedAt, want.CreatedAt)
		}
		if !got.UpdatedAt.After(want.UpdatedAt) {
			t.Errorf("got updatedAt %v, want after %v", got.UpdatedAt, want.UpdatedAt)
		}
		AssertTimestampPrecisionAtLeastMillis(t, pool, "boards", "created_at", "updated_at")

		for _, storedBoard := range ListBoards(t, pool) {
			if storedBoard.ID == got.ID {
				if diff := cmp.Diff(got, storedBoard, testutil.CmpAllowUnexported()); diff != "" {
					t.Errorf("got stored board mismatch (-want +got):\n%s", diff)
				}
				return
			}
		}
		t.Errorf("updated board %v not found", got.ID)
	}

	successUpdateOutboxEvents := func(ownerID domain.UserID, boardName domain.BoardName) []WantOutboxEvent {
		return []WantOutboxEvent{{
			RecipientUserID: ownerID,
			Type:            "board.updated",
			Payload: struct {
				CallerEmail string `json:"callerEmail"`
				BoardName   string `json:"boardName"`
			}{
				CallerEmail: testutil.ValidEmail().String(),
				BoardName:   boardName.String(),
			},
		}}
	}

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

			callerID := fixture.board.OwnerID
			board := fixture.board
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

			want := testutil.UpdateValidBoard(t, &fixture.board, "Updated Board Name", "Updated Board Description", fixture.board.UpdatedAt)
			got, err := r.Update(context.Background(), callerID, board.ID, &want.Name, &want.Description)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Update() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				assertUpdatedBoard(t, got, want)
				AssertOutboxEvents(t, pool, successUpdateOutboxEvents(fixture.board.OwnerID, got.Name))
				return
			}

			if diff := cmp.Diff([]domain.Board{fixture.board, fixture.unrelatedBoard}, ListBoards(t, pool), testutil.CmpAllowUnexported()); diff != "" {
				t.Errorf("stored boards mismatch (-want +got):\n%s", diff)
			}
		})
	}

	t.Run("Success partial name only", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		fixture := setupDefaultBoardHierarchy(t, pool)
		want := testutil.UpdateValidBoard(
			t,
			&fixture.board,
			"Updated Board Name Only",
			fixture.board.Description.String(),
			fixture.board.UpdatedAt,
		)

		got, err := r.Update(context.Background(), fixture.board.OwnerID, fixture.board.ID, &want.Name, nil)
		if err != nil {
			t.Errorf("Update() error = %v", err)
		}
		assertUpdatedBoard(t, got, want)
		AssertOutboxEvents(t, pool, successUpdateOutboxEvents(fixture.board.OwnerID, got.Name))
	})

	t.Run("Success partial description only", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		fixture := setupDefaultBoardHierarchy(t, pool)
		want := testutil.UpdateValidBoard(
			t,
			&fixture.board,
			fixture.board.Name.String(),
			"Updated Board Description Only",
			fixture.board.UpdatedAt,
		)

		got, err := r.Update(context.Background(), fixture.board.OwnerID, fixture.board.ID, nil, &want.Description)
		if err != nil {
			t.Errorf("Update() error = %v", err)
		}
		assertUpdatedBoard(t, got, want)
		AssertOutboxEvents(t, pool, successUpdateOutboxEvents(fixture.board.OwnerID, got.Name))
	})

	t.Run("Success no changes", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		fixture := setupDefaultBoardHierarchy(t, pool)

		got, err := r.Update(context.Background(), fixture.board.OwnerID, fixture.board.ID, nil, nil)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if diff := cmp.Diff(fixture.board, got, testutil.CmpAllowUnexported()); diff != "" {
			t.Errorf("Update() mismatch (-want +got):\n%s", diff)
		}
		AssertOutboxEvents(t, pool, []WantOutboxEvent{})
	})
}

func TestBoardRepository_Delete(t *testing.T) {
	pool, r := boardRepoPrelude(t)

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

			callerID := fixture.board.OwnerID
			board := fixture.board
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

			err := r.Delete(context.Background(), callerID, board.ID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Delete() error = %v, want %v", err, tt.wantErr)
			}

			want := []domain.Board{fixture.board, fixture.unrelatedBoard}
			if tt.wantErr == nil {
				want = []domain.Board{fixture.unrelatedBoard}
				AssertOutboxEvents(t, pool, []WantOutboxEvent{{
					RecipientUserID: board.OwnerID,
					Type:            "board.deleted",
					Payload: struct {
						CallerEmail string `json:"callerEmail"`
						BoardName   string `json:"boardName"`
					}{
						CallerEmail: testutil.ValidEmail().String(),
						BoardName:   board.Name.String(),
					},
				}})
			}
			if diff := cmp.Diff(want, ListBoards(t, pool), testutil.CmpAllowUnexported()); diff != "" {
				t.Errorf("stored boards mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBoardRepository_GetAggregate(t *testing.T) {
	pool, r := boardRepoPrelude(t)

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

			callerID := fixture.board.OwnerID
			board := fixture.board
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

			gotBoard, gotColumns, gotTasks, err := r.GetAggregate(context.Background(), callerID, board.ID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("GetAggregate() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if diff := cmp.Diff(fixture.board, gotBoard, testutil.CmpAllowUnexported()); diff != "" {
					t.Errorf("board mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff([]domain.Column{fixture.column, fixture.siblingColumn}, gotColumns, testutil.CmpAllowUnexported()); diff != "" {
					t.Errorf("columns mismatch (-want +got):\n%s", diff)
				}
				if diff := cmp.Diff([]domain.Task{fixture.task, fixture.parallelTask}, gotTasks, testutil.CmpAllowUnexported()); diff != "" {
					t.Errorf("tasks mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func boardRepoPrelude(t *testing.T) (*pgxpool.Pool, *repository.PGBoard) {
	t.Helper()

	pool := testutil.SetupPostgres(t)
	t.Cleanup(func() { pool.Close() })

	return pool, repository.NewPGBoard(pool)
}
