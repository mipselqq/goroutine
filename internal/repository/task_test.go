//go:build integration

package repository_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/testutil"
)

func TestTaskRepository_Create(t *testing.T) {
	pool, r := taskRepoPrelude(t)

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
			columnWithoutTasks := testutil.NewValidColumn(t, fixture.board.ID, "Empty", 3)
			CreateColumn(t, pool, &columnWithoutTasks)
			validTask := testutil.ValidTask(columnWithoutTasks.ID)

			callerID := fixture.board.OwnerID
			board := fixture.board
			column := columnWithoutTasks
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

			task, err := r.Create(context.Background(), callerID, board.ID, column.ID, validTask.Name, validTask.Description)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Create() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if got := ListTasksByColumnID(t, pool, columnWithoutTasks.ID); len(got) != 0 {
					t.Errorf("got %d tasks, want 0", len(got))
				}
				return
			}

			if task.ID.IsNil() {
				t.Error("got empty task id, want generated id")
			}
			if task.ColumnID != columnWithoutTasks.ID {
				t.Errorf("got columnID %q, want %q", task.ColumnID, columnWithoutTasks.ID)
			}
			if task.Name != validTask.Name {
				t.Errorf("got name %q, want %q", task.Name, validTask.Name)
			}
			if task.Description != validTask.Description {
				t.Errorf("got description %q, want %q", task.Description, validTask.Description)
			}
			if task.Position.Int64() != 1 {
				t.Errorf("got position %d, want 1", task.Position.Int64())
			}
			if task.CreatedAt.IsZero() {
				t.Errorf("got zero createdAt, want set value")
			}
			if task.UpdatedAt.IsZero() {
				t.Errorf("got zero updatedAt, want set value")
			}
			if !task.CreatedAt.Equal(task.UpdatedAt) {
				t.Errorf("got createdAt=%v updatedAt=%v, want equal", task.CreatedAt, task.UpdatedAt)
			}
			AssertTimestampPrecisionAtLeastMillis(t, pool, "tasks", "created_at", "updated_at")

			storedTasks := ListTasksByColumnID(t, pool, columnWithoutTasks.ID)
			if len(storedTasks) != 1 {
				t.Fatalf("ListTasksByColumnID() returned %d tasks, want exactly 1", len(storedTasks))
			}
			if diff := cmp.Diff(task, storedTasks[0], testutil.CmpAllowUnexported()); diff != "" {
				t.Errorf("got stored task mismatch (-want +got):\n%s", diff)
			}

			AssertOutboxEvents(t, pool, []WantOutboxEvent{{
				RecipientUserID: fixture.board.OwnerID,
				Type:            "task.created",
				Payload: struct {
					CallerEmail string `json:"callerEmail"`
					BoardName   string `json:"boardName"`
					ColumnName  string `json:"columnName"`
					TaskName    string `json:"taskName"`
				}{
					CallerEmail: testutil.ValidEmail().String(),
					BoardName:   fixture.board.Name.String(),
					ColumnName:  columnWithoutTasks.Name.String(),
					TaskName:    task.Name.String(),
				},
			}})
		})
	}
}

func TestTaskRepository_Create_AppendsPosition(t *testing.T) {
	pool, r := taskRepoPrelude(t)
	testutil.TruncateAllTables(t, pool)

	fixture := setupDefaultBoardHierarchy(t, pool)
	column := fixture.column

	secondTask := testutil.NewValidTask(t, column.ID, "Second", "Second description", 2)

	createdTask, err := r.Create(
		context.Background(),
		fixture.board.OwnerID,
		fixture.board.ID,
		column.ID,
		secondTask.Name,
		secondTask.Description,
	)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if createdTask.Position.Int64() != 2 {
		t.Errorf("got position %d, want 2", createdTask.Position.Int64())
	}
}

func TestTaskRepository_ListByColumnID(t *testing.T) {
	pool, r := taskRepoPrelude(t)

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
		{name: "Success empty"},
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
			columnWithoutTasks := testutil.NewValidColumn(t, fixture.board.ID, "Empty", 3)
			CreateColumn(t, pool, &columnWithoutTasks)

			callerID := fixture.board.OwnerID
			board := fixture.board
			column := columnWithoutTasks
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

			tasks, err := r.ListByColumnID(context.Background(), callerID, board.ID, column.ID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ListByColumnID() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && len(tasks) != 0 {
				t.Errorf("got %d tasks, want 0", len(tasks))
			}
		})
	}

	t.Run("Success ordered and filtered by column", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		fixture := setupDefaultBoardHierarchy(t, pool)
		board := fixture.board
		firstColumn := fixture.column
		secondTask := testutil.NewValidTask(t, firstColumn.ID, "Second", "second", 2)
		CreateTask(t, pool, &secondTask)

		got, err := r.ListByColumnID(context.Background(), board.OwnerID, board.ID, firstColumn.ID)
		if err != nil {
			t.Fatalf("ListByColumnID() error = %v", err)
		}

		want := []domain.Task{fixture.task, secondTask}
		if diff := cmp.Diff(want, got, testutil.CmpAllowUnexported()); diff != "" {
			t.Errorf("ListByColumnID() mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestTaskRepository_Get(t *testing.T) {
	pool, r := taskRepoPrelude(t)

	tests := []struct {
		name               string
		useUnrelatedOwner  bool
		useMissingOwner    bool
		useUnrelatedBoard  bool
		useMissingBoard    bool
		useSiblingColumn   bool
		useUnrelatedColumn bool
		useMissingColumn   bool
		useParallelTask    bool
		useMissingTask     bool
		wantErr            error
	}{
		{name: "Success"},
		{name: "Unrelated owner", useUnrelatedOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing owner", useMissingOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Unrelated board", useUnrelatedBoard: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing board", useMissingBoard: true, wantErr: repository.ErrRowNotFound},
		{name: "Sibling column", useSiblingColumn: true, wantErr: repository.ErrRowNotFound},
		{name: "Unrelated column", useUnrelatedColumn: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing column", useMissingColumn: true, wantErr: repository.ErrRowNotFound},
		{name: "Task from sibling column", useParallelTask: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing task", useMissingTask: true, wantErr: repository.ErrRowNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.TruncateAllTables(t, pool)
			fixture := setupDefaultBoardHierarchy(t, pool)
			callerID := fixture.board.OwnerID
			board := fixture.board
			column := fixture.column
			task := fixture.task
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
			if tt.useSiblingColumn {
				column = fixture.siblingColumn
			}
			if tt.useUnrelatedColumn {
				column = fixture.unrelatedColumn
			}
			if tt.useMissingColumn {
				column = fixture.nonexistentColumn
			}
			if tt.useParallelTask {
				task = fixture.parallelTask
			}
			if tt.useMissingTask {
				task = fixture.nonexistentTask
			}

			got, err := r.Get(context.Background(), callerID, board.ID, column.ID, task.ID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Get() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if diff := cmp.Diff(fixture.task, got, testutil.CmpAllowUnexported()); diff != "" {
					t.Errorf("Get() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestTaskRepository_Update(t *testing.T) {
	pool, r := taskRepoPrelude(t)

	assertUpdatedTask := func(t *testing.T, got domain.Task, want domain.Task) {
		t.Helper()

		if got.ID != want.ID {
			t.Errorf("got id %q, want %q", got.ID, want.ID)
		}
		if got.ColumnID != want.ColumnID {
			t.Errorf("got columnID %q, want %q", got.ColumnID, want.ColumnID)
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
		AssertTimestampPrecisionAtLeastMillis(t, pool, "tasks", "created_at", "updated_at")

		storedTasks := ListTasksByColumnID(t, pool, want.ColumnID)
		if len(storedTasks) != 1 {
			t.Fatalf("ListTasksByColumnID() returned %d tasks, want exactly 1", len(storedTasks))
		}
		if diff := cmp.Diff(got, storedTasks[0], testutil.CmpAllowUnexported()); diff != "" {
			t.Errorf("got stored task mismatch (-want +got):\n%s", diff)
		}
	}

	successUpdateOutboxEvents := func(ownerID domain.UserID, taskName domain.TaskName) []WantOutboxEvent {
		return []WantOutboxEvent{{
			RecipientUserID: ownerID,
			Type:            "task.updated",
			Payload: struct {
				CallerEmail string `json:"callerEmail"`
				BoardName   string `json:"boardName"`
				ColumnName  string `json:"columnName"`
				TaskName    string `json:"taskName"`
			}{
				CallerEmail: testutil.ValidEmail().String(),
				BoardName:   testutil.ValidBoardName().String(),
				ColumnName:  testutil.ValidColumn(domain.NewBoardID()).Name.String(),
				TaskName:    taskName.String(),
			},
		}}
	}

	tests := []struct {
		name               string
		useUnrelatedOwner  bool
		useMissingOwner    bool
		useUnrelatedBoard  bool
		useMissingBoard    bool
		useSiblingColumn   bool
		useUnrelatedColumn bool
		useMissingColumn   bool
		useParalleltask    bool
		useMissingTask     bool
		wantErr            error
	}{
		{name: "Success"},
		{name: "Unrelated owner", useUnrelatedOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing owner", useMissingOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Unrelated board", useUnrelatedBoard: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing board", useMissingBoard: true, wantErr: repository.ErrRowNotFound},
		{name: "Sibling column", useSiblingColumn: true, wantErr: repository.ErrRowNotFound},
		{name: "Unrelated column", useUnrelatedColumn: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing column", useMissingColumn: true, wantErr: repository.ErrRowNotFound},
		{name: "Task from sibling column", useParalleltask: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing task", useMissingTask: true, wantErr: repository.ErrRowNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.TruncateAllTables(t, pool)
			fixture := setupDefaultBoardHierarchy(t, pool)
			want := testutil.UpdateValidTask(t, &fixture.task, "Renamed", "Renamed description", fixture.task.UpdatedAt)

			callerID := fixture.board.OwnerID
			board := fixture.board
			column := fixture.column
			task := fixture.task
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
			if tt.useSiblingColumn {
				column = fixture.siblingColumn
			}
			if tt.useUnrelatedColumn {
				column = fixture.unrelatedColumn
			}
			if tt.useMissingColumn {
				column = fixture.nonexistentColumn
			}
			if tt.useParalleltask {
				task = fixture.parallelTask
			}
			if tt.useMissingTask {
				task = fixture.nonexistentTask
			}

			got, err := r.Update(context.Background(), callerID, board.ID, column.ID, task.ID, &want.Name, &want.Description)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Update() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				assertUpdatedTask(t, got, want)
				AssertOutboxEvents(t, pool, successUpdateOutboxEvents(fixture.board.OwnerID, got.Name))
				return
			}

			if diff := cmp.Diff([]domain.Task{fixture.task}, ListTasksByColumnID(t, pool, fixture.column.ID), testutil.CmpAllowUnexported()); diff != "" {
				t.Errorf("stored tasks mismatch (-want +got):\n%s", diff)
			}
		})
	}

	t.Run("Success partial name only", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		fixture := setupDefaultBoardHierarchy(t, pool)
		want := testutil.UpdateValidTask(
			t,
			&fixture.task,
			"Renamed name only",
			fixture.task.Description.String(),
			fixture.task.UpdatedAt,
		)

		got, err := r.Update(
			context.Background(),
			fixture.board.OwnerID,
			fixture.board.ID,
			fixture.column.ID,
			fixture.task.ID,
			&want.Name,
			nil,
		)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		assertUpdatedTask(t, got, want)
		AssertOutboxEvents(t, pool, successUpdateOutboxEvents(fixture.board.OwnerID, got.Name))
	})

	t.Run("Success partial description only", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		fixture := setupDefaultBoardHierarchy(t, pool)
		want := testutil.UpdateValidTask(
			t,
			&fixture.task,
			fixture.task.Name.String(),
			"Renamed description only",
			fixture.task.UpdatedAt,
		)

		got, err := r.Update(
			context.Background(),
			fixture.board.OwnerID,
			fixture.board.ID,
			fixture.column.ID,
			fixture.task.ID,
			nil,
			&want.Description,
		)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		assertUpdatedTask(t, got, want)
		AssertOutboxEvents(t, pool, successUpdateOutboxEvents(fixture.board.OwnerID, got.Name))
	})

	t.Run("Success no changes", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		fixture := setupDefaultBoardHierarchy(t, pool)

		got, err := r.Update(
			context.Background(),
			fixture.board.OwnerID,
			fixture.board.ID,
			fixture.column.ID,
			fixture.task.ID,
			nil,
			nil,
		)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if diff := cmp.Diff(fixture.task, got, testutil.CmpAllowUnexported()); diff != "" {
			t.Errorf("Update() mismatch (-want +got):\n%s", diff)
		}
		AssertOutboxEvents(t, pool, []WantOutboxEvent{})
	})
}

func TestTaskRepository_Move(t *testing.T) {
	pool, r := taskRepoPrelude(t)

	tests := []struct {
		name                          string
		useUnrelatedOwner             bool
		useMissingOwner               bool
		useUnrelatedBoard             bool
		useMissingBoard               bool
		useSiblingSourceColumn        bool
		useUnrelatedSourceColumn      bool
		useMissingSourceColumn        bool
		useParallelTask               bool
		useMissingTask                bool
		useUnrelatedDestinationColumn bool
		useMissingDestinationColumn   bool
		wantErr                       error
	}{
		{name: "Success move down within column"},
		{name: "Unrelated owner", useUnrelatedOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing owner", useMissingOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Unrelated board", useUnrelatedBoard: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing board", useMissingBoard: true, wantErr: repository.ErrRowNotFound},
		{name: "Sibling source column", useSiblingSourceColumn: true, wantErr: repository.ErrRowNotFound},
		{name: "Unrelated source column", useUnrelatedSourceColumn: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing source column", useMissingSourceColumn: true, wantErr: repository.ErrRowNotFound},
		{name: "Task from sibling column", useParallelTask: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing task", useMissingTask: true, wantErr: repository.ErrRowNotFound},
		{name: "Unrelated destination column", useUnrelatedDestinationColumn: true, wantErr: repository.ErrTargetRowNotFound},
		{name: "Missing destination column", useMissingDestinationColumn: true, wantErr: repository.ErrTargetRowNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.TruncateAllTables(t, pool)
			fixture := setupDefaultBoardHierarchy(t, pool)
			secondTask := testutil.NewValidTask(t, fixture.column.ID, "Second", "second", 2)
			thirdTask := testutil.NewValidTask(t, fixture.column.ID, "Third", "third", 3)
			CreateTask(t, pool, &thirdTask)
			CreateTask(t, pool, &secondTask)
			callerID := fixture.board.OwnerID
			board := fixture.board
			sourceColumn := fixture.column
			task := fixture.task
			destinationColumn := fixture.column
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
			if tt.useSiblingSourceColumn {
				sourceColumn = fixture.siblingColumn
			}
			if tt.useUnrelatedSourceColumn {
				sourceColumn = fixture.unrelatedColumn
			}
			if tt.useMissingSourceColumn {
				sourceColumn = fixture.nonexistentColumn
			}
			if tt.useParallelTask {
				task = fixture.parallelTask
			}
			if tt.useMissingTask {
				task = fixture.nonexistentTask
			}
			if tt.useUnrelatedDestinationColumn {
				destinationColumn = fixture.unrelatedColumn
			}
			if tt.useMissingDestinationColumn {
				destinationColumn = fixture.nonexistentColumn
			}

			destinationPosition := testutil.NewValidTaskPosition(t, 3)
			gotColumn, gotPosition, err := r.Move(
				context.Background(),
				callerID,
				board.ID,
				sourceColumn.ID,
				task.ID,
				destinationColumn.ID,
				destinationPosition,
			)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Move() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if gotColumn != fixture.column.ID {
					t.Errorf("Move() column = %v, want %v", gotColumn, fixture.column.ID)
				}
				if gotPosition != destinationPosition {
					t.Errorf("Move() position = %v, want %v", gotPosition, destinationPosition)
				}
			}

			got := ListTasksByColumnID(t, pool, fixture.column.ID)
			if len(got) != 3 {
				t.Fatalf("got %d tasks after move, want 3", len(got))
			}
			if tt.wantErr == nil {
				assertTaskIDAndPosition(t, &got[0], secondTask.ID, 1)
				assertTaskIDAndPosition(t, &got[1], thirdTask.ID, 2)
				assertTaskIDAndPosition(t, &got[2], fixture.task.ID, 3)
				return
			}
			assertTaskIDAndPosition(t, &got[0], fixture.task.ID, 1)
			assertTaskIDAndPosition(t, &got[1], secondTask.ID, 2)
			assertTaskIDAndPosition(t, &got[2], thirdTask.ID, 3)
		})
	}

	t.Run("Success move up within column", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		fixture := setupDefaultBoardHierarchy(t, pool)
		board := fixture.board
		sourceColumn := fixture.column
		secondTask := testutil.NewValidTask(t, sourceColumn.ID, "Second", "second", 2)
		thirdTask := testutil.NewValidTask(t, sourceColumn.ID, "Third", "third", 3)

		CreateTask(t, pool, &secondTask)
		CreateTask(t, pool, &thirdTask)

		destinationPosition := testutil.NewValidTaskPosition(t, 1)

		gotColumn, gotPosition, err := r.Move(context.Background(), board.OwnerID, board.ID, sourceColumn.ID, thirdTask.ID, sourceColumn.ID, destinationPosition)
		if err != nil {
			t.Fatalf("Move() error = %v", err)
		}
		if gotColumn != sourceColumn.ID {
			t.Fatalf("Move() column = %v, want %v", gotColumn, sourceColumn.ID)
		}
		if gotPosition != destinationPosition {
			t.Fatalf("Move() position = %v, want %v", gotPosition, destinationPosition)
		}

		got := ListTasksByColumnID(t, pool, sourceColumn.ID)
		if len(got) != 3 {
			t.Fatalf("got %d tasks after move, want 3", len(got))
		}
		assertTaskIDAndPosition(t, &got[0], thirdTask.ID, 1)
		assertTaskIDAndPosition(t, &got[1], fixture.task.ID, 2)
		assertTaskIDAndPosition(t, &got[2], secondTask.ID, 3)
	})

	t.Run("Success no-op", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		fixture := setupDefaultBoardHierarchy(t, pool)
		board := fixture.board
		sourceColumn := fixture.column
		secondTask := testutil.NewValidTask(t, sourceColumn.ID, "Second", "second", 2)

		CreateTask(t, pool, &secondTask)

		destinationPosition := testutil.NewValidTaskPosition(t, 2)

		gotColumn, gotPosition, err := r.Move(context.Background(), board.OwnerID, board.ID, sourceColumn.ID, secondTask.ID, sourceColumn.ID, destinationPosition)
		if err != nil {
			t.Fatalf("Move() error = %v", err)
		}
		if gotColumn != sourceColumn.ID {
			t.Fatalf("Move() column = %v, want %v", gotColumn, sourceColumn.ID)
		}
		if gotPosition != destinationPosition {
			t.Fatalf("Move() position = %v, want %v", gotPosition, destinationPosition)
		}

		got := ListTasksByColumnID(t, pool, sourceColumn.ID)
		if len(got) != 2 {
			t.Fatalf("got %d tasks after no-op move, want 2", len(got))
		}
		assertTaskIDAndPosition(t, &got[0], fixture.task.ID, 1)
		assertTaskIDAndPosition(t, &got[1], secondTask.ID, 2)
		AssertOutboxEvents(t, pool, []WantOutboxEvent{})
	})

	t.Run("Index out of bounds within column", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		fixture := setupDefaultBoardHierarchy(t, pool)
		board := fixture.board
		sourceColumn := fixture.column
		secondTask := testutil.NewValidTask(t, sourceColumn.ID, "Second", "second", 2)
		thirdTask := testutil.NewValidTask(t, sourceColumn.ID, "Third", "third", 3)

		CreateTask(t, pool, &secondTask)
		CreateTask(t, pool, &thirdTask)

		destinationPosition := testutil.NewValidTaskPosition(t, 4)

		_, _, err := r.Move(context.Background(), board.OwnerID, board.ID, sourceColumn.ID, secondTask.ID, sourceColumn.ID, destinationPosition)
		if !errors.Is(err, repository.ErrIndexOutOfBounds) {
			t.Fatalf("Move() error = %v, want ErrIndexOutOfBounds", err)
		}

		got := ListTasksByColumnID(t, pool, sourceColumn.ID)
		if len(got) != 3 {
			t.Fatalf("got %d tasks after failed move, want 3", len(got))
		}
		assertTaskIDAndPosition(t, &got[0], fixture.task.ID, 1)
		assertTaskIDAndPosition(t, &got[1], secondTask.ID, 2)
		assertTaskIDAndPosition(t, &got[2], thirdTask.ID, 3)
	})

	t.Run("Success move across columns into middle", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		fixture := setupDefaultBoardHierarchy(t, pool)
		board := fixture.board
		sourceColumn := fixture.column
		firstTask := fixture.task
		destinationColumn := fixture.siblingColumn

		secondTask := testutil.NewValidTask(t, sourceColumn.ID, "A2", "a2", 2)
		thirdTask := testutil.NewValidTask(t, sourceColumn.ID, "A3", "a3", 3)

		firstDestinationTask := fixture.parallelTask
		secondDestinationTask := testutil.NewValidTask(t, destinationColumn.ID, "B2", "b2", 2)

		CreateTask(t, pool, &thirdTask)
		CreateTask(t, pool, &secondTask)
		CreateTask(t, pool, &secondDestinationTask)

		destinationPosition := testutil.NewValidTaskPosition(t, 2)

		gotColumn, gotPosition, err := r.Move(context.Background(), board.OwnerID, board.ID, sourceColumn.ID, secondTask.ID, destinationColumn.ID, destinationPosition)
		if err != nil {
			t.Fatalf("Move() error = %v", err)
		}
		if gotColumn != destinationColumn.ID {
			t.Fatalf("Move() column = %v, want %v", gotColumn, destinationColumn.ID)
		}
		if gotPosition != destinationPosition {
			t.Fatalf("Move() position = %v, want %v", gotPosition, destinationPosition)
		}

		sourceTasks := ListTasksByColumnID(t, pool, sourceColumn.ID)
		if len(sourceTasks) != 2 {
			t.Fatalf("got %d tasks in source column after move, want 2", len(sourceTasks))
		}
		assertTaskIDAndPosition(t, &sourceTasks[0], firstTask.ID, 1)
		assertTaskIDAndPosition(t, &sourceTasks[1], thirdTask.ID, 2)

		destinationTasks := ListTasksByColumnID(t, pool, destinationColumn.ID)
		if len(destinationTasks) != 3 {
			t.Fatalf("got %d tasks in destination column after move, want 3", len(destinationTasks))
		}
		assertTaskIDAndPosition(t, &destinationTasks[0], firstDestinationTask.ID, 1)
		assertTaskIDAndPosition(t, &destinationTasks[1], secondTask.ID, 2)
		assertTaskIDAndPosition(t, &destinationTasks[2], secondDestinationTask.ID, 3)

		AssertOutboxEvents(t, pool, []WantOutboxEvent{{
			RecipientUserID: board.OwnerID,
			Type:            "task.moved",
			Payload: struct {
				CallerEmail      string `json:"callerEmail"`
				BoardName        string `json:"boardName"`
				TaskName         string `json:"taskName"`
				SourceColumnName string `json:"sourceColumnName"`
				TargetColumnName string `json:"targetColumnName"`
				SourcePosition   int64  `json:"sourcePosition"`
				TargetPosition   int64  `json:"targetPosition"`
			}{
				CallerEmail:      testutil.ValidEmail().String(),
				BoardName:        board.Name.String(),
				TaskName:         secondTask.Name.String(),
				SourceColumnName: sourceColumn.Name.String(),
				TargetColumnName: destinationColumn.Name.String(),
				SourcePosition:   secondTask.Position.Int64(),
				TargetPosition:   destinationPosition.Int64(),
			},
		}})
	})

	t.Run("Success move across columns to append", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		fixture := setupDefaultBoardHierarchy(t, pool)
		board := fixture.board
		sourceColumn := fixture.column
		firstTask := fixture.task
		destinationColumn := fixture.siblingColumn
		firstDestinationTask := fixture.parallelTask

		destinationPosition := testutil.NewValidTaskPosition(t, 2)

		gotColumn, gotPosition, err := r.Move(context.Background(), board.OwnerID, board.ID, sourceColumn.ID, firstTask.ID, destinationColumn.ID, destinationPosition)
		if err != nil {
			t.Fatalf("Move() error = %v", err)
		}
		if gotColumn != destinationColumn.ID {
			t.Fatalf("Move() column = %v, want %v", gotColumn, destinationColumn.ID)
		}
		if gotPosition != destinationPosition {
			t.Fatalf("Move() position = %v, want %v", gotPosition, destinationPosition)
		}

		sourceTasks := ListTasksByColumnID(t, pool, sourceColumn.ID)
		if len(sourceTasks) != 0 {
			t.Fatalf("got %d tasks in source column after move, want 0", len(sourceTasks))
		}

		destinationTasks := ListTasksByColumnID(t, pool, destinationColumn.ID)
		if len(destinationTasks) != 2 {
			t.Fatalf("got %d tasks in destination column after move, want 2", len(destinationTasks))
		}
		assertTaskIDAndPosition(t, &destinationTasks[0], firstDestinationTask.ID, 1)
		assertTaskIDAndPosition(t, &destinationTasks[1], firstTask.ID, 2)
	})

	t.Run("Index out of bounds across columns", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		fixture := setupDefaultBoardHierarchy(t, pool)
		board := fixture.board
		sourceColumn := fixture.column
		firstTask := fixture.task
		destinationColumn := fixture.siblingColumn
		firstDestinationTask := fixture.parallelTask

		destinationPosition := testutil.NewValidTaskPosition(t, 3)

		_, _, err := r.Move(context.Background(), board.OwnerID, board.ID, sourceColumn.ID, firstTask.ID, destinationColumn.ID, destinationPosition)
		if !errors.Is(err, repository.ErrIndexOutOfBounds) {
			t.Fatalf("Move() error = %v, want ErrIndexOutOfBounds", err)
		}

		sourceTasks := ListTasksByColumnID(t, pool, sourceColumn.ID)
		if len(sourceTasks) != 1 {
			t.Fatalf("got %d tasks in source column after failed move, want 1", len(sourceTasks))
		}
		assertTaskIDAndPosition(t, &sourceTasks[0], firstTask.ID, 1)

		destinationTasks := ListTasksByColumnID(t, pool, destinationColumn.ID)
		if len(destinationTasks) != 1 {
			t.Fatalf("got %d tasks in destination column after failed move, want 1", len(destinationTasks))
		}
		assertTaskIDAndPosition(t, &destinationTasks[0], firstDestinationTask.ID, 1)
	})
}

func TestTaskRepository_Delete(t *testing.T) {
	pool, r := taskRepoPrelude(t)

	tests := []struct {
		name               string
		useUnrelatedOwner  bool
		useMissingOwner    bool
		useUnrelatedBoard  bool
		useMissingBoard    bool
		useSiblingColumn   bool
		useUnrelatedColumn bool
		useMissingColumn   bool
		useParallelTask    bool
		useMissingTask     bool
		wantErr            error
	}{
		{name: "Success shift positions"},
		{name: "Unrelated owner", useUnrelatedOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing owner", useMissingOwner: true, wantErr: repository.ErrRowNotFound},
		{name: "Unrelated board", useUnrelatedBoard: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing board", useMissingBoard: true, wantErr: repository.ErrRowNotFound},
		{name: "Sibling column", useSiblingColumn: true, wantErr: repository.ErrRowNotFound},
		{name: "Unrelated column", useUnrelatedColumn: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing column", useMissingColumn: true, wantErr: repository.ErrRowNotFound},
		{name: "Task from sibling column", useParallelTask: true, wantErr: repository.ErrRowNotFound},
		{name: "Missing task", useMissingTask: true, wantErr: repository.ErrRowNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testutil.TruncateAllTables(t, pool)
			fixture := setupDefaultBoardHierarchy(t, pool)
			secondTask := testutil.NewValidTask(t, fixture.column.ID, "Second", "second", 2)
			thirdTask := testutil.NewValidTask(t, fixture.column.ID, "Third", "third", 3)
			CreateTask(t, pool, &thirdTask)
			CreateTask(t, pool, &secondTask)
			callerID := fixture.board.OwnerID
			board := fixture.board
			column := fixture.column
			task := secondTask
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
			if tt.useSiblingColumn {
				column = fixture.siblingColumn
			}
			if tt.useUnrelatedColumn {
				column = fixture.unrelatedColumn
			}
			if tt.useMissingColumn {
				column = fixture.nonexistentColumn
			}
			if tt.useParallelTask {
				task = fixture.parallelTask
			}
			if tt.useMissingTask {
				task = fixture.nonexistentTask
			}

			err := r.Delete(context.Background(), callerID, board.ID, column.ID, task.ID)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Delete() error = %v, want %v", err, tt.wantErr)
			}

			got := ListTasksByColumnID(t, pool, fixture.column.ID)
			if tt.wantErr == nil {
				if len(got) != 2 {
					t.Fatalf("got %d tasks after delete, want 2", len(got))
				}
				assertTaskIDAndPosition(t, &got[0], fixture.task.ID, 1)
				assertTaskIDAndPosition(t, &got[1], thirdTask.ID, 2)

				AssertOutboxEvents(t, pool, []WantOutboxEvent{{
					RecipientUserID: fixture.board.OwnerID,
					Type:            "task.deleted",
					Payload: struct {
						CallerEmail string `json:"callerEmail"`
						BoardName   string `json:"boardName"`
						ColumnName  string `json:"columnName"`
						TaskName    string `json:"taskName"`
					}{
						CallerEmail: testutil.ValidEmail().String(),
						BoardName:   fixture.board.Name.String(),
						ColumnName:  fixture.column.Name.String(),
						TaskName:    secondTask.Name.String(),
					},
				}})
				return
			}
			if len(got) != 3 {
				t.Fatalf("got %d tasks after failed delete, want 3", len(got))
			}
			assertTaskIDAndPosition(t, &got[0], fixture.task.ID, 1)
			assertTaskIDAndPosition(t, &got[1], secondTask.ID, 2)
			assertTaskIDAndPosition(t, &got[2], thirdTask.ID, 3)
		})
	}
}

func TestLockTaskColumns_BlocksSecondTransaction(t *testing.T) {
	pool, _ := taskRepoPrelude(t)
	testutil.TruncateAllTables(t, pool)

	fixture := setupDefaultBoardHierarchy(t, pool)

	beginTx := func(id string) pgx.Tx {
		tx, err := pool.Begin(context.Background())
		if err != nil {
			t.Fatalf("pool.Begin() tx%s error = %v", id, err)
		}
		return tx
	}

	setLockTimeoutMs := func(tx pgx.Tx, id string, ms int) {
		_, err := tx.Exec(context.Background(), fmt.Sprintf(`SET LOCAL lock_timeout = '%dms'`, ms))
		if err != nil {
			t.Fatalf("tx%s SET LOCAL lock_timeout error = %v", id, err)
		}
	}

	rollbackTx := func(tx pgx.Tx, id string) {
		err := tx.Rollback(context.Background())
		if err != nil {
			t.Fatalf("tx%s Rollback() error = %v", id, err)
		}
	}

	lockTaskColumns := func(tx pgx.Tx) error {
		_, err := repository.LockTaskColumns(context.Background(), tx, fixture.board.OwnerID, fixture.board.ID, fixture.column.ID)
		return err
	}

	tx1 := beginTx("1")

	// 1. LockTaskColumns runs SELECT ... FOR UPDATE on the columns row for this board/column;
	//    that row lock blocks any other transaction from locking the same row until tx1 ends.
	err := lockTaskColumns(tx1)
	if err != nil {
		t.Fatalf("LockTaskColumns() tx1 error = %v", err)
	}

	tx2 := beginTx("2")

	// 2. tx2: the next LockTaskColumns will try the same FOR UPDATE on the same row. While tx1
	//    still holds the lock, PostgreSQL would wait forever; SET LOCAL lock_timeout limits that
	//    wait to ~100ms, then the statement fails with a lock timeout instead of hanging the test.
	setLockTimeoutMs(tx2, "2", 100)

	// 3. tx2 must fail to acquire the same lock while tx1 still holds it.
	err = lockTaskColumns(tx2)
	if err == nil {
		t.Fatal("second LockTaskColumns() unexpectedly succeeded while tx1 still held the lock")
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Fatalf("second LockTaskColumns: want wrapped *pgconn.PgError, got %T: %v", err, err)
	}
	if pgErr.Code != "55P03" {
		t.Fatalf("second LockTaskColumns: want SQLSTATE 55P03 (lock_not_available because of lock timeout), got %v", err)
	}

	// 4. Roll back tx2 after the lock wait timeout.
	rollbackTx(tx2, "2")
	// 5. Roll back tx1 to release the original lock.
	rollbackTx(tx1, "1")

	// 6. Start tx3 after the lock has been released.
	tx3 := beginTx("3")

	// 7. tx3 should now acquire the same lock successfully.
	err = lockTaskColumns(tx3)
	if err != nil {
		t.Fatalf("third LockTaskColumns() after release error = %v", err)
	}

	// 8. Clean up tx3.
	rollbackTx(tx3, "3")
}

func assertTaskIDAndPosition(t *testing.T, task *domain.Task, wantID domain.TaskID, wantPos int64) {
	t.Helper()

	if task.ID != wantID {
		t.Errorf("got id %q, want %q", task.ID, wantID)
	}
	if task.Position.Int64() != wantPos {
		t.Errorf("got position %d, want %d", task.Position.Int64(), wantPos)
	}
}

func taskRepoPrelude(t *testing.T) (*pgxpool.Pool, *repository.PGTask) {
	t.Helper()

	pool := testutil.SetupPostgres(t)
	t.Cleanup(func() { pool.Close() })

	return pool, repository.NewPGTask(pool)
}
