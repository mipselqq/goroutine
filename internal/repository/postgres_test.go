package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/testutil"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateFixedUser(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	id := testutil.ValidUserID()
	domainEmail := testutil.ValidEmail()
	CreateUser(t, pool, id, domainEmail, testutil.ValidPasswordHash())
}

func CreateUser(t *testing.T, pool *pgxpool.Pool, id domain.UserID, email domain.Email, hash domain.PasswordHash) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const query = `INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)`
	_, err := pool.Exec(ctx, query, id, email, hash.RevealSecret())
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
}

func CreateBoard(t *testing.T, pool *pgxpool.Pool, board *domain.Board) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const q = `
			INSERT INTO boards (id, owner_id, name, description, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := pool.Exec(
		ctx, q,
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

	const q = `
			SELECT id, owner_id, name, description, created_at, updated_at
			FROM boards
			ORDER BY created_at ASC`

	rows, err := pool.Query(ctx, q)
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

	const q = `
			INSERT INTO columns (id, board_id, name, description, position, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := pool.Exec(
		ctx, q,
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

	const q = `
			SELECT id, board_id, name, description, position, created_at, updated_at
			FROM columns
			WHERE board_id = $1
			ORDER BY position ASC`

	rows, err := pool.Query(ctx, q, boardID)
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

	const q = `
			INSERT INTO tasks (id, column_id, name, description, position, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := pool.Exec(
		ctx, q,
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

func ListTasksByColumnID(t *testing.T, pool *pgxpool.Pool, columnID domain.ColumnID) []domain.Task {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const q = `
			SELECT id, column_id, name, description, position, created_at, updated_at
			FROM tasks
			WHERE column_id = $1
			ORDER BY position ASC`

	rows, err := pool.Query(ctx, q, columnID)
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

func insertFixedUserAndBoard(t *testing.T, pool *pgxpool.Pool) domain.Board {
	t.Helper()

	CreateFixedUser(t, pool)
	board := testutil.ValidBoard()
	CreateBoard(t, pool, &board)

	return board
}

func insertFixedUserBoardAndColumn(t *testing.T, pool *pgxpool.Pool) (domain.Board, domain.Column) {
	t.Helper()

	board := insertFixedUserAndBoard(t, pool)
	column := testutil.ValidColumn(board.ID)
	CreateColumn(t, pool, &column)

	return board, column
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

func ListUsers(t *testing.T, pool *pgxpool.Pool) []domain.User {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const q = `SELECT id, email, password_hash, telegram_chat_id, telegram_username FROM users ORDER BY id`

	rows, err := pool.Query(ctx, q)
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
