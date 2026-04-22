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

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		_, column := insertFixedUserBoardAndColumn(t, pool)

		validTask := testutil.ValidTask(column.ID)

		task, err := r.Create(
			context.Background(),
			column.ID,
			validTask.Name,
			validTask.Description,
		)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if task.ID.IsEmpty() {
			t.Error("got empty task id, want generated id")
		}
		if task.ColumnID != column.ID {
			t.Errorf("got columnID %q, want %q", task.ColumnID, column.ID)
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

		stored, ok := FindTaskByID(t, pool, task.ID)
		if !ok {
			t.Fatalf("created task %q not found in DB", task.ID)
		}
		if diff := cmp.Diff(task, stored, testutil.CmpAllowUnexported()); diff != "" {
			t.Errorf("got stored task mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestTaskRepository_Create_AppendsPosition(t *testing.T) {
	pool, r := taskRepoPrelude(t)

	testutil.TruncateAllTables(t, pool)

	_, column := insertFixedUserBoardAndColumn(t, pool)

	existing := testutil.ValidTask(column.ID)
	InsertTask(t, pool, &existing)

	toCreate := testutil.NewValidTask(t, column.ID, "Second", "Second description", 2)

	second, err := r.Create(
		context.Background(),
		column.ID,
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

func TestTaskRepository_ListByColumnID(t *testing.T) {
	pool, r := taskRepoPrelude(t)

	t.Run("Success empty", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		_, column := insertFixedUserBoardAndColumn(t, pool)

		tasks, err := r.ListByColumnID(context.Background(), column.ID)
		if err != nil {
			t.Fatalf("ListByColumnID() error = %v", err)
		}
		if len(tasks) != 0 {
			t.Fatalf("got %d tasks, want 0", len(tasks))
		}
	})

	t.Run("Success ordered and filtered by column", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board, columnA := insertFixedUserBoardAndColumn(t, pool)
		columnB := testutil.NewValidColumn(t, board.ID, "In Progress", 2)
		InsertColumn(t, pool, &columnB)

		first := testutil.ValidTask(columnA.ID)
		second := testutil.NewValidTask(t, columnA.ID, "Second", "second", 2)
		otherColumnTask := testutil.NewValidTask(t, columnB.ID, "Other", "other", 1)

		InsertTask(t, pool, &first)
		InsertTask(t, pool, &second)
		InsertTask(t, pool, &otherColumnTask)

		got, err := r.ListByColumnID(context.Background(), columnA.ID)
		if err != nil {
			t.Fatalf("ListByColumnID() error = %v", err)
		}

		want := []domain.Task{first, second}
		if diff := cmp.Diff(want, got, testutil.CmpAllowUnexported()); diff != "" {
			t.Errorf("ListByColumnID() mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestTaskRepository_GetByID(t *testing.T) {
	pool, r := taskRepoPrelude(t)

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		_, column := insertFixedUserBoardAndColumn(t, pool)

		created := testutil.ValidTask(column.ID)
		InsertTask(t, pool, &created)

		got, err := r.GetByID(context.Background(), created.ID)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if diff := cmp.Diff(created, got, testutil.CmpAllowUnexported()); diff != "" {
			t.Errorf("GetByID() mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("Not found", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		_, err := r.GetByID(context.Background(), domain.NewTaskID())
		assertErrRowNotFound(t, err)
	})
}

func TestTaskRepository_UpdateByID(t *testing.T) {
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

		stored, ok := FindTaskByID(t, pool, want.ID)
		if !ok {
			t.Fatalf("updated task %q not found in DB", want.ID)
		}
		if diff := cmp.Diff(got, stored, testutil.CmpAllowUnexported()); diff != "" {
			t.Errorf("got stored task mismatch (-want +got):\n%s", diff)
		}
	}

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		_, column := insertFixedUserBoardAndColumn(t, pool)

		created := testutil.ValidTask(column.ID)
		createdAtBeforeUpdate := time.Now().UTC()
		updatedAtBeforeUpdate := createdAtBeforeUpdate
		created.CreatedAt = createdAtBeforeUpdate
		created.UpdatedAt = updatedAtBeforeUpdate
		InsertTask(t, pool, &created)

		want := testutil.UpdateValidTask(t, &created, "Renamed", "Renamed description", testutil.FixedTime5mFromNow())
		updated, err := r.UpdateByID(context.Background(), column.ID, created.ID, &want.Name, &want.Description)
		if err != nil {
			t.Fatalf("UpdateByID() error = %v", err)
		}

		assertUpdatedTask(t, updated, want)
	})

	t.Run("Not found by task id", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		_, column := insertFixedUserBoardAndColumn(t, pool)

		updatedName, _ := domain.NewTaskName("Renamed")
		_, err := r.UpdateByID(context.Background(), column.ID, domain.NewTaskID(), &updatedName, nil)
		assertErrRowNotFound(t, err)
	})

	t.Run("Not found by column id", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		_, column := insertFixedUserBoardAndColumn(t, pool)

		created := testutil.ValidTask(column.ID)
		InsertTask(t, pool, &created)

		want := testutil.UpdateValidTask(t, &created, "Renamed", "Renamed description", testutil.FixedTime5mFromNow())
		_, err := r.UpdateByID(context.Background(), domain.NewColumnID(), created.ID, &want.Name, &want.Description)
		assertErrRowNotFound(t, err)
	})
}

func TestTaskRepository_Move(t *testing.T) {
	pool, r := taskRepoPrelude(t)

	t.Run("Success move down within column", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board, column := insertFixedUserBoardAndColumn(t, pool)

		first := testutil.ValidTask(column.ID)
		second := testutil.NewValidTask(t, column.ID, "Second", "second", 2)
		third := testutil.NewValidTask(t, column.ID, "Third", "third", 3)

		InsertTask(t, pool, &first)
		InsertTask(t, pool, &second)
		InsertTask(t, pool, &third)

		targetPosition := testutil.MustTaskPosition(t, 3)

		gotColumn, gotPosition, err := r.Move(context.Background(), board.ID, column.ID, first.ID, column.ID, targetPosition)
		if err != nil {
			t.Fatalf("Move() error = %v", err)
		}
		if gotColumn != column.ID {
			t.Fatalf("Move() column = %v, want %v", gotColumn, column.ID)
		}
		if gotPosition != targetPosition {
			t.Fatalf("Move() position = %v, want %v", gotPosition, targetPosition)
		}

		got := ListTasksByColumnID(t, pool, column.ID)
		if len(got) != 3 {
			t.Fatalf("got %d tasks after move, want 3", len(got))
		}
		assertTaskIDAndPosition(t, &got[0], second.ID, 1)
		assertTaskIDAndPosition(t, &got[1], third.ID, 2)
		assertTaskIDAndPosition(t, &got[2], first.ID, 3)
	})

	t.Run("Success move up within column", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board, column := insertFixedUserBoardAndColumn(t, pool)

		first := testutil.ValidTask(column.ID)
		second := testutil.NewValidTask(t, column.ID, "Second", "second", 2)
		third := testutil.NewValidTask(t, column.ID, "Third", "third", 3)

		InsertTask(t, pool, &first)
		InsertTask(t, pool, &second)
		InsertTask(t, pool, &third)

		targetPosition := testutil.MustTaskPosition(t, 1)

		gotColumn, gotPosition, err := r.Move(context.Background(), board.ID, column.ID, third.ID, column.ID, targetPosition)
		if err != nil {
			t.Fatalf("Move() error = %v", err)
		}
		if gotColumn != column.ID {
			t.Fatalf("Move() column = %v, want %v", gotColumn, column.ID)
		}
		if gotPosition != targetPosition {
			t.Fatalf("Move() position = %v, want %v", gotPosition, targetPosition)
		}

		got := ListTasksByColumnID(t, pool, column.ID)
		if len(got) != 3 {
			t.Fatalf("got %d tasks after move, want 3", len(got))
		}
		assertTaskIDAndPosition(t, &got[0], third.ID, 1)
		assertTaskIDAndPosition(t, &got[1], first.ID, 2)
		assertTaskIDAndPosition(t, &got[2], second.ID, 3)
	})

	t.Run("Success no-op", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board, column := insertFixedUserBoardAndColumn(t, pool)

		first := testutil.ValidTask(column.ID)
		second := testutil.NewValidTask(t, column.ID, "Second", "second", 2)

		InsertTask(t, pool, &first)
		InsertTask(t, pool, &second)

		targetPosition := testutil.MustTaskPosition(t, 2)

		gotColumn, gotPosition, err := r.Move(context.Background(), board.ID, column.ID, second.ID, column.ID, targetPosition)
		if err != nil {
			t.Fatalf("Move() error = %v", err)
		}
		if gotColumn != column.ID {
			t.Fatalf("Move() column = %v, want %v", gotColumn, column.ID)
		}
		if gotPosition != targetPosition {
			t.Fatalf("Move() position = %v, want %v", gotPosition, targetPosition)
		}

		got := ListTasksByColumnID(t, pool, column.ID)
		if len(got) != 2 {
			t.Fatalf("got %d tasks after no-op move, want 2", len(got))
		}
		assertTaskIDAndPosition(t, &got[0], first.ID, 1)
		assertTaskIDAndPosition(t, &got[1], second.ID, 2)
	})

	t.Run("Index out of bounds within column", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board, column := insertFixedUserBoardAndColumn(t, pool)

		first := testutil.ValidTask(column.ID)
		second := testutil.NewValidTask(t, column.ID, "Second", "second", 2)
		third := testutil.NewValidTask(t, column.ID, "Third", "third", 3)

		InsertTask(t, pool, &first)
		InsertTask(t, pool, &second)
		InsertTask(t, pool, &third)

		targetPosition := testutil.MustTaskPosition(t, 4)

		_, _, err := r.Move(context.Background(), board.ID, column.ID, second.ID, column.ID, targetPosition)
		if !errors.Is(err, repository.ErrIndexOutOfBounds) {
			t.Fatalf("Move() error = %v, want ErrIndexOutOfBounds", err)
		}

		got := ListTasksByColumnID(t, pool, column.ID)
		if len(got) != 3 {
			t.Fatalf("got %d tasks after failed move, want 3", len(got))
		}
		assertTaskIDAndPosition(t, &got[0], first.ID, 1)
		assertTaskIDAndPosition(t, &got[1], second.ID, 2)
		assertTaskIDAndPosition(t, &got[2], third.ID, 3)
	})

	t.Run("Success move across columns into middle", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board, columnA := insertFixedUserBoardAndColumn(t, pool)
		columnB := testutil.NewValidColumn(t, board.ID, "In Progress", 2)
		InsertColumn(t, pool, &columnB)

		a1 := testutil.ValidTask(columnA.ID)
		a2 := testutil.NewValidTask(t, columnA.ID, "A2", "a2", 2)
		a3 := testutil.NewValidTask(t, columnA.ID, "A3", "a3", 3)

		b1 := testutil.NewValidTask(t, columnB.ID, "B1", "b1", 1)
		b2 := testutil.NewValidTask(t, columnB.ID, "B2", "b2", 2)

		InsertTask(t, pool, &a1)
		InsertTask(t, pool, &a2)
		InsertTask(t, pool, &a3)
		InsertTask(t, pool, &b1)
		InsertTask(t, pool, &b2)

		targetPosition := testutil.MustTaskPosition(t, 2)

		gotColumn, gotPosition, err := r.Move(context.Background(), board.ID, columnA.ID, a2.ID, columnB.ID, targetPosition)
		if err != nil {
			t.Fatalf("Move() error = %v", err)
		}
		if gotColumn != columnB.ID {
			t.Fatalf("Move() column = %v, want %v", gotColumn, columnB.ID)
		}
		if gotPosition != targetPosition {
			t.Fatalf("Move() position = %v, want %v", gotPosition, targetPosition)
		}

		gotA := ListTasksByColumnID(t, pool, columnA.ID)
		if len(gotA) != 2 {
			t.Fatalf("got %d tasks in source column after move, want 2", len(gotA))
		}
		assertTaskIDAndPosition(t, &gotA[0], a1.ID, 1)
		assertTaskIDAndPosition(t, &gotA[1], a3.ID, 2)

		gotB := ListTasksByColumnID(t, pool, columnB.ID)
		if len(gotB) != 3 {
			t.Fatalf("got %d tasks in target column after move, want 3", len(gotB))
		}
		assertTaskIDAndPosition(t, &gotB[0], b1.ID, 1)
		assertTaskIDAndPosition(t, &gotB[1], a2.ID, 2)
		assertTaskIDAndPosition(t, &gotB[2], b2.ID, 3)

		movedTask, ok := FindTaskByID(t, pool, a2.ID)
		if !ok {
			t.Fatalf("moved task %q not found in DB", a2.ID)
		}
		if movedTask.ColumnID != columnB.ID {
			t.Errorf("got moved columnID %q, want %q", movedTask.ColumnID, columnB.ID)
		}
	})

	t.Run("Success move across columns to append", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board, columnA := insertFixedUserBoardAndColumn(t, pool)
		columnB := testutil.NewValidColumn(t, board.ID, "Done", 2)
		InsertColumn(t, pool, &columnB)

		a1 := testutil.ValidTask(columnA.ID)
		b1 := testutil.NewValidTask(t, columnB.ID, "B1", "b1", 1)

		InsertTask(t, pool, &a1)
		InsertTask(t, pool, &b1)

		targetPosition := testutil.MustTaskPosition(t, 2)

		gotColumn, gotPosition, err := r.Move(context.Background(), board.ID, columnA.ID, a1.ID, columnB.ID, targetPosition)
		if err != nil {
			t.Fatalf("Move() error = %v", err)
		}
		if gotColumn != columnB.ID {
			t.Fatalf("Move() column = %v, want %v", gotColumn, columnB.ID)
		}
		if gotPosition != targetPosition {
			t.Fatalf("Move() position = %v, want %v", gotPosition, targetPosition)
		}

		gotA := ListTasksByColumnID(t, pool, columnA.ID)
		if len(gotA) != 0 {
			t.Fatalf("got %d tasks in source column after move, want 0", len(gotA))
		}

		gotB := ListTasksByColumnID(t, pool, columnB.ID)
		if len(gotB) != 2 {
			t.Fatalf("got %d tasks in target column after move, want 2", len(gotB))
		}
		assertTaskIDAndPosition(t, &gotB[0], b1.ID, 1)
		assertTaskIDAndPosition(t, &gotB[1], a1.ID, 2)
	})

	t.Run("Index out of bounds across columns", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board, columnA := insertFixedUserBoardAndColumn(t, pool)
		columnB := testutil.NewValidColumn(t, board.ID, "Done", 2)
		InsertColumn(t, pool, &columnB)

		a1 := testutil.ValidTask(columnA.ID)
		b1 := testutil.NewValidTask(t, columnB.ID, "B1", "b1", 1)

		InsertTask(t, pool, &a1)
		InsertTask(t, pool, &b1)

		targetPosition := testutil.MustTaskPosition(t, 3)

		_, _, err := r.Move(context.Background(), board.ID, columnA.ID, a1.ID, columnB.ID, targetPosition)
		if !errors.Is(err, repository.ErrIndexOutOfBounds) {
			t.Fatalf("Move() error = %v, want ErrIndexOutOfBounds", err)
		}

		gotA := ListTasksByColumnID(t, pool, columnA.ID)
		if len(gotA) != 1 {
			t.Fatalf("got %d tasks in source column after failed move, want 1", len(gotA))
		}
		assertTaskIDAndPosition(t, &gotA[0], a1.ID, 1)

		gotB := ListTasksByColumnID(t, pool, columnB.ID)
		if len(gotB) != 1 {
			t.Fatalf("got %d tasks in target column after failed move, want 1", len(gotB))
		}
		assertTaskIDAndPosition(t, &gotB[0], b1.ID, 1)
	})

	t.Run("Not found by task id", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board, column := insertFixedUserBoardAndColumn(t, pool)

		targetPosition := testutil.MustTaskPosition(t, 1)

		_, _, err := r.Move(context.Background(), board.ID, column.ID, domain.NewTaskID(), column.ID, targetPosition)
		assertErrRowNotFound(t, err)
	})

	t.Run("Not found by board id", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		_, column := insertFixedUserBoardAndColumn(t, pool)

		created := testutil.ValidTask(column.ID)
		InsertTask(t, pool, &created)

		targetPosition := testutil.MustTaskPosition(t, 1)

		_, _, err := r.Move(context.Background(), domain.NewBoardID(), column.ID, created.ID, column.ID, targetPosition)
		assertErrRowNotFound(t, err)
	})
}

func TestTaskRepository_Delete(t *testing.T) {
	pool, r := taskRepoPrelude(t)

	t.Run("Success shift positions", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board, column := insertFixedUserBoardAndColumn(t, pool)

		first := testutil.ValidTask(column.ID)
		second := testutil.NewValidTask(t, column.ID, "Second", "second", 2)
		third := testutil.NewValidTask(t, column.ID, "Third", "third", 3)

		InsertTask(t, pool, &first)
		InsertTask(t, pool, &second)
		InsertTask(t, pool, &third)

		err := r.Delete(context.Background(), board.ID, column.ID, second.ID)
		if err != nil {
			t.Fatalf("Delete() error = %v", err)
		}

		got := ListTasksByColumnID(t, pool, column.ID)

		if len(got) != 2 {
			t.Fatalf("got %d tasks after delete, want 2", len(got))
		}
		assertTaskIDAndPosition(t, &got[0], first.ID, 1)
		assertTaskIDAndPosition(t, &got[1], third.ID, 2)

		_, ok := FindTaskByID(t, pool, second.ID)
		if ok {
			t.Error("got deleted task in DB, want absent")
		}
	})

	t.Run("Not found by task id", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		board, column := insertFixedUserBoardAndColumn(t, pool)

		err := r.Delete(context.Background(), board.ID, column.ID, domain.NewTaskID())
		assertErrRowNotFound(t, err)
	})

	t.Run("Not found by board id", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		_, column := insertFixedUserBoardAndColumn(t, pool)

		created := testutil.ValidTask(column.ID)
		InsertTask(t, pool, &created)

		err := r.Delete(context.Background(), domain.NewBoardID(), column.ID, created.ID)
		assertErrRowNotFound(t, err)
	})
}

func TestLockTaskColumns_BlocksSecondTransaction(t *testing.T) {
	pool, _ := taskRepoPrelude(t)
	testutil.TruncateAllTables(t, pool)

	board, column := insertFixedUserBoardAndColumn(t, pool)

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
		return repository.LockTaskColumns(context.Background(), tx, board.ID, column.ID)
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

func taskRepoPrelude(t *testing.T) (*pgxpool.Pool, *repository.PgTask) {
	t.Helper()

	pool := testutil.SetupTestDB(t, "../../migrations")
	t.Cleanup(func() { pool.Close() })

	return pool, repository.NewPgTask(pool)
}
