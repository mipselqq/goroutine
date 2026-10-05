package repository_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/testutil"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateUser(t *testing.T, pool *pgxpool.Pool, id domain.UserID, email domain.Email, hash domain.PasswordHash) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const query = `
		INSERT INTO users (id, email, password_hash)
		VALUES ($1, $2, $3)`
	_, err := pool.Exec(ctx, query, id, email, hash.RevealSecret())
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
}

func LinkTelegramChat(t *testing.T, pool *pgxpool.Pool, userID domain.UserID) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const query = `
		UPDATE users
		SET telegram_chat_id = $1,
			telegram_username = $2
		WHERE id = $3`
	_, err := pool.Exec(ctx, query, testutil.ValidTelegramChatID(), testutil.ValidTelegramUsername(), userID)
	if err != nil {
		t.Fatalf("LinkTelegramChat() error = %v", err)
	}
}

func CreateBoard(t *testing.T, pool *pgxpool.Pool, board *domain.Board) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const query = `
		INSERT INTO boards (id, owner_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := pool.Exec(
		ctx, query,
		board.ID,
		board.OwnerID,
		board.Name,
		board.Description,
		board.CreatedAt,
		board.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("CreateBoard() error = %v", err)
	}
}

func ListBoards(t *testing.T, pool *pgxpool.Pool) []domain.Board {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const query = `
		SELECT id, owner_id, name, description, created_at, updated_at
		FROM boards
		ORDER BY created_at ASC`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		t.Fatalf("ListBoards() error = %v", err)
	}
	defer rows.Close()

	var boards []domain.Board
	for rows.Next() {
		board, scanErr := repository.ScanBoard(rows)
		if scanErr != nil {
			t.Fatalf("Board row Scan() error = %v", scanErr)
		}
		boards = append(boards, board)
	}

	err = rows.Err()
	if err != nil {
		t.Fatalf("rows.Err() error = %v", err)
	}

	return boards
}

func CreateColumn(t *testing.T, pool *pgxpool.Pool, column *domain.Column) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const query = `
		INSERT INTO columns (id, board_id, name, description, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := pool.Exec(
		ctx, query,
		column.ID,
		column.BoardID,
		column.Name,
		column.Description,
		column.Position,
		column.CreatedAt,
		column.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("CreateColumn() error = %v", err)
	}
}

func ListColumnsByBoardID(t *testing.T, pool *pgxpool.Pool, boardID domain.BoardID) []domain.Column {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const query = `
		SELECT id, board_id, name, description, position, created_at, updated_at
		FROM columns
		WHERE board_id = $1
		ORDER BY position ASC`

	rows, err := pool.Query(ctx, query, boardID)
	if err != nil {
		t.Fatalf("ListColumnsByBoardID() error = %v", err)
	}
	defer rows.Close()

	var columns []domain.Column
	for rows.Next() {
		column, scanErr := repository.ScanColumn(rows)
		if scanErr != nil {
			t.Fatalf("Column row Scan() error = %v", scanErr)
		}

		columns = append(columns, column)
	}

	err = rows.Err()
	if err != nil {
		t.Fatalf("rows.Err() error = %v", err)
	}

	return columns
}

func CreateTask(t *testing.T, pool *pgxpool.Pool, task *domain.Task) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const query = `
		INSERT INTO tasks (id, column_id, name, description, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := pool.Exec(
		ctx, query,
		task.ID,
		task.ColumnID,
		task.Name,
		task.Description,
		task.Position,
		task.CreatedAt,
		task.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}
}

type boardHierarchyFixture struct {
	board             domain.Board
	column            domain.Column
	task              domain.Task
	siblingColumn     domain.Column
	parallelTask      domain.Task
	unrelatedBoard    domain.Board
	unrelatedColumn   domain.Column
	unrelatedTask     domain.Task
	nonexistentBoard  domain.Board
	nonexistentColumn domain.Column
	nonexistentTask   domain.Task
}

// setupDefaultBoardHierarchy creates this fixture:
//
// board
// ├── column -> task
// └── siblingColumn -> parallelTask
// unrelatedBoard
// └── unrelatedColumn -> unrelatedTask
// nonexistentBoard
// └── nonexistentColumn -> nonexistentTask
//
// All entities except 'nonexistent' are inserted into the database.
func setupDefaultBoardHierarchy(t *testing.T, pool *pgxpool.Pool) boardHierarchyFixture {
	t.Helper()

	board := testutil.ValidBoard()
	column := testutil.ValidColumn(board.ID)
	task := testutil.ValidTask(column.ID)
	siblingColumn := testutil.NewValidColumn(t, board.ID, "Sibling", 2)
	parallelTask := testutil.ValidTask(siblingColumn.ID)

	unrelatedBoard := testutil.ValidBoardForOwner(domain.NewUserID())
	unrelatedColumn := testutil.ValidColumn(unrelatedBoard.ID)
	unrelatedTask := testutil.ValidTask(unrelatedColumn.ID)

	CreateUser(t, pool, board.OwnerID, testutil.ValidEmail(), testutil.ValidPasswordHash())
	CreateUser(t, pool, unrelatedBoard.OwnerID, testutil.AnotherValidEmail(), testutil.ValidPasswordHash())
	CreateBoard(t, pool, &board)
	CreateColumn(t, pool, &column)
	CreateTask(t, pool, &task)
	CreateColumn(t, pool, &siblingColumn)
	CreateTask(t, pool, &parallelTask)
	CreateBoard(t, pool, &unrelatedBoard)
	CreateColumn(t, pool, &unrelatedColumn)
	CreateTask(t, pool, &unrelatedTask)

	nonexistentBoard := testutil.ValidBoardForOwner(domain.NewUserID())
	nonexistentColumn := testutil.ValidColumn(nonexistentBoard.ID)
	nonexistentTask := testutil.ValidTask(nonexistentColumn.ID)

	return boardHierarchyFixture{
		board:             board,
		column:            column,
		task:              task,
		siblingColumn:     siblingColumn,
		parallelTask:      parallelTask,
		unrelatedBoard:    unrelatedBoard,
		unrelatedColumn:   unrelatedColumn,
		unrelatedTask:     unrelatedTask,
		nonexistentBoard:  nonexistentBoard,
		nonexistentColumn: nonexistentColumn,
		nonexistentTask:   nonexistentTask,
	}
}

func ListTasksByColumnID(t *testing.T, pool *pgxpool.Pool, columnID domain.ColumnID) []domain.Task {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const query = `
		SELECT id, column_id, name, description, position, created_at, updated_at
		FROM tasks
		WHERE column_id = $1
		ORDER BY position ASC`

	rows, err := pool.Query(ctx, query, columnID)
	if err != nil {
		t.Fatalf("ListTasksByColumnID() error = %v", err)
	}
	defer rows.Close()

	var tasks []domain.Task
	for rows.Next() {
		task, scanErr := repository.ScanTask(rows)
		if scanErr != nil {
			t.Fatalf("Task row Scan() error = %v", scanErr)
		}

		tasks = append(tasks, task)
	}

	err = rows.Err()
	if err != nil {
		t.Fatalf("rows.Err() error = %v", err)
	}

	return tasks
}

func AssertTimestampPrecisionAtLeastMillis(t *testing.T, pool *pgxpool.Pool, tableName string, columnNames ...string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const query = `
		SELECT datetime_precision
		FROM information_schema.columns
		WHERE table_schema = current_schema()
		    AND table_name = $1
		    AND column_name = $2`

	for _, columnName := range columnNames {
		var precision int32
		err := pool.QueryRow(ctx, query, tableName, columnName).Scan(&precision)
		if err != nil {
			t.Fatalf("timestamp precision lookup for %s.%s: %v", tableName, columnName, err)
		}
		if precision < 3 {
			t.Errorf("got datetime_precision=%d for %s.%s, want >= 3", precision, tableName, columnName)
		}
	}
}

type WantOutboxEvent struct {
	RecipientUserID domain.UserID
	Type            string
	Payload         any
}

func AssertOutboxEvents(t *testing.T, pool *pgxpool.Pool, wantEvents []WantOutboxEvent) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const query = `
		SELECT id, recipient_user_id, event_type, payload, created_at
		FROM notification_outbox
		ORDER BY id`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		t.Fatalf("query outbox events: %v", err)
	}
	defer rows.Close()

	eventsCount := 0
	for rows.Next() {
		var (
			gotID           int64
			gotRecipientID  uuid.UUID
			gotType         string
			gotPayloadBytes []byte
			gotCreatedAt    time.Time
		)
		err = rows.Scan(&gotID, &gotRecipientID, &gotType, &gotPayloadBytes, &gotCreatedAt)
		if err != nil {
			t.Fatalf("scan outbox event %d: %v", eventsCount, err)
		}
		eventsCount++

		if eventsCount > len(wantEvents) {
			t.Errorf("got unexpected extra outbox event %d: recipient_user_id=%s event_type=%q payload=%s, want %d", eventsCount, gotRecipientID, gotType, gotPayloadBytes, len(wantEvents))
			continue
		}

		if gotID != int64(eventsCount) {
			t.Errorf("got outbox event %d id=%d, want %d", eventsCount, gotID, eventsCount)
		}
		now := time.Now().UTC()
		if gotCreatedAt.Before(now.Add(15 * -time.Second)) {
			t.Errorf("got outbox event %d created_at=%v at least 15s before check time %v, want not too late", eventsCount, gotCreatedAt, now)
		}
		if gotCreatedAt.After(now) {
			t.Errorf("got outbox event %d created_at=%v after check time %v, want not in the future", eventsCount, gotCreatedAt, now)
		}

		wantEvent := wantEvents[eventsCount-1]
		if gotRecipientID != wantEvent.RecipientUserID.UUID() {
			t.Errorf("outbox event %d recipient_user_id=%s, want %s", eventsCount, gotRecipientID, wantEvent.RecipientUserID)
		}
		if gotType != wantEvent.Type {
			t.Errorf("outbox event %d event_type=%q, want %q", eventsCount, gotType, wantEvent.Type)
		}

		var gotPayload any
		err = json.Unmarshal(gotPayloadBytes, &gotPayload)
		if err != nil {
			t.Fatalf("unmarshal outbox event %d payload: %v", eventsCount, err)
		}

		expectedPayloadBytes, marshalErr := json.Marshal(wantEvent.Payload)
		if marshalErr != nil {
			t.Fatalf("marshal expected outbox event %d payload: %v", eventsCount, marshalErr)
		}
		var wantPayload any
		err = json.Unmarshal(expectedPayloadBytes, &wantPayload)
		if err != nil {
			t.Fatalf("normalize expected outbox event %d payload: %v", eventsCount, marshalErr)
		}

		diff := cmp.Diff(wantPayload, gotPayload)
		if diff != "" {
			t.Errorf("outbox event %d payload mismatch (-want +got):\n%s", eventsCount, diff)
		}

	}

	err = rows.Err()
	if err != nil {
		t.Fatalf("iterate outbox events: %v", err)
	}
	if eventsCount < len(wantEvents) {
		t.Errorf("got %d outbox events, want %d", eventsCount, len(wantEvents))
	}

	AssertTimestampPrecisionAtLeastMillis(t, pool, "notification_outbox", "created_at")
}

func ListUsers(t *testing.T, pool *pgxpool.Pool) []domain.User {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const query = `
		SELECT id, email, password_hash, telegram_chat_id, telegram_username
		FROM users
		ORDER BY id`

	rows, err := pool.Query(ctx, query)
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		user, scanErr := repository.ScanUser(rows)
		if scanErr != nil {
			t.Fatalf("User row Scan() error = %v", scanErr)
		}
		users = append(users, user)
	}

	err = rows.Err()
	if err != nil {
		t.Fatalf("rows.Err() error = %v", err)
	}

	return users
}

func assertErrRowNotFound(t *testing.T, err error) {
	t.Helper()

	if !errors.Is(err, repository.ErrRowNotFound) {
		t.Errorf("got error %v, want ErrRowNotFound", err)
	}
}
