//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/testutil"
)

func TestBoardRepository_Create(t *testing.T) {
	pool := testutil.SetupTestDB(t, "../../migrations")
	defer pool.Close()

	r := repository.NewPgBoard(pool)
	userID := testutil.ValidUserID()
	boardName := testutil.ValidBoardName()
	boardDescription := testutil.ValidBoardDescription()

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		CreateUser(t, pool, userID, "test@example.com")

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

		const query = `
		SELECT owner_id, name, description, created_at, updated_at 
		FROM boards 
		WHERE id = $1`

		var (
			dbOwnerID     domain.UserID
			dbName        domain.BoardName
			dbDescription domain.BoardDescription
			dbCreatedAt   time.Time
			dbUpdatedAt   time.Time
		)
		err = pool.QueryRow(context.Background(), query, board.ID).
			Scan(&dbOwnerID, &dbName, &dbDescription, &dbCreatedAt, &dbUpdatedAt)
		if err != nil {
			t.Fatalf("Board row Scan() error = %v", err)
		}
		if dbOwnerID != userID {
			t.Errorf("DB: got owner ID %q, want %q", dbOwnerID, userID)
		}
		if dbName != boardName {
			t.Errorf("DB: got name %q, want %q", dbName, boardName)
		}
		if dbDescription != boardDescription {
			t.Errorf("DB: got description %q, want %q", dbDescription, boardDescription)
		}
		if !dbCreatedAt.Equal(board.CreatedAt) {
			t.Errorf("DB: got created_at %v, want %v", dbCreatedAt, board.CreatedAt)
		}
		if !dbUpdatedAt.Equal(board.UpdatedAt) {
			t.Errorf("DB: got updated_at %v, want %v", dbUpdatedAt, board.UpdatedAt)
		}
	})
}

func TestBoardRepository_GetByID(t *testing.T) {
	pool := testutil.SetupTestDB(t, "../../migrations")
	defer pool.Close()

	r := repository.NewPgBoard(pool)
	userID := testutil.ValidUserID()
	boardName := testutil.ValidBoardName()
	boardDescription := testutil.ValidBoardDescription()

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		CreateUser(t, pool, userID, "getbyid@example.com")

		now := time.Now().UTC()
		want := domain.Board{
			ID:          domain.NewBoardID(),
			OwnerID:     userID,
			Name:        boardName,
			Description: boardDescription,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		InsertBoard(t, pool, &want)

		got, err := r.GetByID(context.Background(), want.ID)
		if err != nil {
			t.Errorf("GetByID() error = %v", err)
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

		_, err := r.GetByID(context.Background(), domain.NewBoardID())
		assertErrRowNotFound(t, err)
	})
}

func TestBoardRepository_GetMany(t *testing.T) {
	pool := testutil.SetupTestDB(t, "../../migrations")
	defer pool.Close()

	r := repository.NewPgBoard(pool)
	userID := testutil.ValidUserID()
	boardName := testutil.ValidBoardName()
	boardDescription := testutil.ValidBoardDescription()

	t.Run("Success empty", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		CreateUser(t, pool, userID, "getmany-empty@example.com")

		got, err := r.GetMany(context.Background(), userID)
		if err != nil {
			t.Errorf("GetMany() error = %v", err)
		}
		if len(got) != 0 {
			t.Errorf("got %d boards, want 0", len(got))
		}
	})

	t.Run("Success returns boards in created order", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		CreateUser(t, pool, userID, "getmany-order@example.com")

		otherName, err := domain.NewBoardName(boardName.String() + "-2")
		if err != nil {
			t.Fatalf("NewBoardName() error = %v", err)
		}
		first := domain.Board{
			ID:          domain.NewBoardID(),
			OwnerID:     userID,
			Name:        boardName,
			Description: boardDescription,
			CreatedAt:   testutil.FixedTimeNow(),
			UpdatedAt:   testutil.FixedTimeNow(),
		}
		second := domain.Board{
			ID:          domain.NewBoardID(),
			OwnerID:     userID,
			Name:        otherName,
			Description: boardDescription,
			CreatedAt:   testutil.FixedTime5mFromNow(),
			UpdatedAt:   testutil.FixedTime5mFromNow(),
		}

		InsertBoard(t, pool, &first)
		InsertBoard(t, pool, &second)

		got, err := r.GetMany(context.Background(), userID)
		if err != nil {
			t.Errorf("GetMany() error = %v", err)
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
	pool := testutil.SetupTestDB(t, "../../migrations")
	defer pool.Close()

	r := repository.NewPgBoard(pool)
	userID := testutil.ValidUserID()

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

		stored, ok := FindBoardByID(t, pool, validBoard.ID)
		if !ok {
			t.Fatalf("updated board %q not found in DB", validBoard.ID)
		}
		if diff := cmp.Diff(got, stored, testutil.CmpAllowUnexported()); diff != "" {
			t.Errorf("got stored board mismatch (-want +got):\n%s", diff)
		}
	}

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		CreateUser(t, pool, userID, "updatebyid@example.com")
		InsertBoard(t, pool, &validBoard)

		got, err := r.UpdateByID(context.Background(), validBoard.ID, &updatedName, &updatedDescription)
		if err != nil {
			t.Errorf("UpdateByID() error = %v", err)
		}
		assertUpdatedBoard(t, got, updatedValidBoard)
	})

	t.Run("Success partial name only", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		CreateUser(t, pool, userID, "updatebyid-partial-name@example.com")
		InsertBoard(t, pool, &validBoard)

		got, err := r.UpdateByID(context.Background(), validBoard.ID, &updatedNameOnly, nil)
		if err != nil {
			t.Errorf("UpdateByID() error = %v", err)
		}
		assertUpdatedBoard(t, got, updatedNameOnlyBoard)
	})

	t.Run("Success partial description only", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		CreateUser(t, pool, userID, "updatebyid-partial-description@example.com")
		InsertBoard(t, pool, &validBoard)

		got, err := r.UpdateByID(context.Background(), validBoard.ID, nil, &updatedDescriptionOnly)
		if err != nil {
			t.Errorf("UpdateByID() error = %v", err)
		}
		assertUpdatedBoard(t, got, updatedDescriptionOnlyBoard)
	})

	t.Run("Not found when missing", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		CreateUser(t, pool, userID, "updatebyid-missing@example.com")

		_, err := r.UpdateByID(context.Background(), domain.NewBoardID(), &updatedName, &updatedDescription)
		assertErrRowNotFound(t, err)
	})
}

func TestBoardRepository_Delete(t *testing.T) {
	pool := testutil.SetupTestDB(t, "../../migrations")
	defer pool.Close()

	r := repository.NewPgBoard(pool)
	userID := testutil.ValidUserID()

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board := insertFixedUserAndBoard(t, pool, "delete@example.com")

		err := r.Delete(context.Background(), board.ID)
		if err != nil {
			t.Errorf("Delete() error = %v", err)
		}

		_, ok := FindBoardByID(t, pool, board.ID)
		if ok {
			t.Errorf("got board %q in DB, want deleted row", board.ID)
		}
	})

	t.Run("Not found when missing", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		CreateUser(t, pool, userID, "delete-missing@example.com")

		err := r.Delete(context.Background(), domain.NewBoardID())
		assertErrRowNotFound(t, err)
	})
}
