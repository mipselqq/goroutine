//go:build integration

package repository_test

import (
	"context"
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

	userID := testutil.ValidUserID()
	boardName := testutil.ValidBoardName()
	boardDescription := testutil.ValidBoardDescription()

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		CreateFixedUser(t, pool)

		board, err := r.Create(context.Background(), userID, boardName, boardDescription)
		if err != nil {
			t.Errorf("Create() error = %v", err)
		}
		if board.ID.IsEmpty() {
			t.Errorf("got empty board ID, want generated ID")
		}
		if board.OwnerID != userID {
			t.Errorf("got owner ID %q, want %q", board.OwnerID, userID)
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

		stored, ok := GetBoardByID(t, pool, board.ID)
		if !ok {
			t.Fatalf("created board %q not found in DB", board.ID)
		}
		if stored.OwnerID != userID {
			t.Errorf("DB: got owner ID %q, want %q", stored.OwnerID, userID)
		}
		if stored.Name != boardName {
			t.Errorf("DB: got name %q, want %q", stored.Name, boardName)
		}
		if stored.Description != boardDescription {
			t.Errorf("DB: got description %q, want %q", stored.Description, boardDescription)
		}
		if !stored.CreatedAt.Equal(board.CreatedAt) {
			t.Errorf("DB: got created_at %v, want %v", stored.CreatedAt, board.CreatedAt)
		}
		if !stored.UpdatedAt.Equal(board.UpdatedAt) {
			t.Errorf("DB: got updated_at %v, want %v", stored.UpdatedAt, board.UpdatedAt)
		}
	})
}

func TestBoardRepository_GetByID(t *testing.T) {
	pool, r := boardRepoPrelude(t)

	userID := testutil.ValidUserID()
	boardName := testutil.ValidBoardName()
	boardDescription := testutil.ValidBoardDescription()

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		CreateFixedUser(t, pool)

		now := time.Now().UTC()
		want := domain.Board{
			ID:          domain.NewBoardID(),
			OwnerID:     userID,
			Name:        boardName,
			Description: boardDescription,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		CreateBoard(t, pool, &want)

		got, err := r.Get(context.Background(), want.ID)
		if err != nil {
			t.Errorf("Get() error = %v", err)
		}
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
		if !got.UpdatedAt.Truncate(time.Millisecond).Equal(want.UpdatedAt.Truncate(time.Millisecond)) {
			t.Errorf("got updatedAt %v, want %v (at millisecond precision)", got.UpdatedAt, want.UpdatedAt)
		}
	})

	t.Run("Not found", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		_, err := r.Get(context.Background(), domain.NewBoardID())
		assertErrRowNotFound(t, err)
	})
}

func TestBoardRepository_GetMany(t *testing.T) {
	pool, r := boardRepoPrelude(t)

	userID := testutil.ValidUserID()
	boardName := testutil.ValidBoardName()
	boardDescription := testutil.ValidBoardDescription()

	t.Run("Success empty", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		CreateFixedUser(t, pool)

		got, err := r.List(context.Background(), userID)
		if err != nil {
			t.Errorf("List() error = %v", err)
		}
		if len(got) != 0 {
			t.Errorf("got %d boards, want 0", len(got))
		}
	})

	t.Run("Success returns boards in created order", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		CreateFixedUser(t, pool)

		otherName, err := domain.NewBoardName(boardName.String() + "-2")
		if err != nil {
			t.Fatalf("NewBoardName() error = %v", err)
		}
		first := domain.Board{
			ID:          domain.NewBoardID(),
			OwnerID:     userID,
			Name:        boardName,
			Description: boardDescription,
			CreatedAt:   testutil.FixedNow(),
			UpdatedAt:   testutil.FixedNow(),
		}
		second := domain.Board{
			ID:          domain.NewBoardID(),
			OwnerID:     userID,
			Name:        otherName,
			Description: boardDescription,
			CreatedAt:   testutil.Fixed5mFromNow(),
			UpdatedAt:   testutil.Fixed5mFromNow(),
		}

		CreateBoard(t, pool, &second)
		CreateBoard(t, pool, &first)

		got, err := r.List(context.Background(), userID)
		if err != nil {
			t.Errorf("List() error = %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("got %d boards, want 2", len(got))
		}
		want := []domain.Board{first, second}
		if diff := cmp.Diff(want, got, testutil.CmpAllowUnexported()); diff != "" {
			t.Errorf("GetMany() mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestBoardRepository_UpdateByID(t *testing.T) {
	pool, r := boardRepoPrelude(t)

	validBoard := testutil.ValidBoard()
	createdAtBeforeUpdate := time.Now().UTC()
	updatedAtBeforeUpdate := createdAtBeforeUpdate
	validBoard.CreatedAt = createdAtBeforeUpdate
	validBoard.UpdatedAt = updatedAtBeforeUpdate
	updatedValidBoard := testutil.UpdateValidBoard(t, &validBoard, "Updated Board Name", "Updated Board Description", validBoard.UpdatedAt)
	updatedNameOnlyBoard := testutil.UpdateValidBoard(t, &validBoard, "Updated Board Name Only", validBoard.Description.String(), validBoard.UpdatedAt)
	updatedDescriptionOnlyBoard := testutil.UpdateValidBoard(t, &validBoard, validBoard.Name.String(), "Updated Board Description Only", validBoard.UpdatedAt)
	updatedName := updatedValidBoard.Name
	updatedDescription := updatedValidBoard.Description
	updatedNameOnly := updatedNameOnlyBoard.Name
	updatedDescriptionOnly := updatedDescriptionOnlyBoard.Description

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

		stored, ok := GetBoardByID(t, pool, validBoard.ID)
		if !ok {
			t.Fatalf("updated board %q not found in DB", validBoard.ID)
		}
		if diff := cmp.Diff(got, stored, testutil.CmpAllowUnexported()); diff != "" {
			t.Errorf("got stored board mismatch (-want +got):\n%s", diff)
		}
	}

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		CreateFixedUser(t, pool)
		CreateBoard(t, pool, &validBoard)

		got, err := r.Update(context.Background(), validBoard.ID, &updatedName, &updatedDescription)
		if err != nil {
			t.Errorf("Update() error = %v", err)
		}
		assertUpdatedBoard(t, got, updatedValidBoard)
	})

	t.Run("Success partial name only", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		CreateFixedUser(t, pool)
		CreateBoard(t, pool, &validBoard)

		got, err := r.Update(context.Background(), validBoard.ID, &updatedNameOnly, nil)
		if err != nil {
			t.Errorf("Update() error = %v", err)
		}
		assertUpdatedBoard(t, got, updatedNameOnlyBoard)
	})

	t.Run("Success partial description only", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		CreateFixedUser(t, pool)
		CreateBoard(t, pool, &validBoard)

		got, err := r.Update(context.Background(), validBoard.ID, nil, &updatedDescriptionOnly)
		if err != nil {
			t.Errorf("Update() error = %v", err)
		}
		assertUpdatedBoard(t, got, updatedDescriptionOnlyBoard)
	})

	t.Run("Not found when missing", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		CreateFixedUser(t, pool)

		_, err := r.Update(context.Background(), domain.NewBoardID(), &updatedName, &updatedDescription)
		assertErrRowNotFound(t, err)
	})
}

func TestBoardRepository_Delete(t *testing.T) {
	pool, r := boardRepoPrelude(t)

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board := insertFixedUserAndBoard(t, pool)

		err := r.Delete(context.Background(), board.ID)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}

		_, ok := GetBoardByID(t, pool, board.ID)
		if ok {
			t.Errorf("got board %q in DB, want deleted row", board.ID)
		}
	})

	t.Run("Not found when missing", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		CreateFixedUser(t, pool)

		err := r.Delete(context.Background(), domain.NewBoardID())
		assertErrRowNotFound(t, err)
	})
}

func boardRepoPrelude(t *testing.T) (*pgxpool.Pool, *repository.PGBoard) {
	t.Helper()

	pool := testutil.SetupPostgres(t, "../../migrations")
	t.Cleanup(func() { pool.Close() })

	return pool, repository.NewPGBoard(pool)
}
