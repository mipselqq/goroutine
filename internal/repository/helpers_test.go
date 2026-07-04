package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/testutil"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const telegramTokenPrefix = "tg_token:"

// TODO: remove this function. Too implicit.
func CreateFixedUser(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	id := testutil.ValidUserID()
	domainEmail := testutil.ValidEmail()
	CreateUser(t, pool, id, domainEmail, testutil.ValidPasswordHash())
}

func CreateUser(t *testing.T, pool *pgxpool.Pool, id domain.UserID, email domain.Email, hash string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const query = `INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)`
	_, err := pool.Exec(ctx, query, id, email, hash)
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

func GetBoardByID(t *testing.T, pool *pgxpool.Pool, boardID domain.BoardID) (domain.Board, bool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const q = `
		SELECT id, owner_id, name, description, created_at, updated_at
		FROM boards
		WHERE id = $1`

	var board domain.Board
	err := pool.QueryRow(ctx, q, boardID).Scan(
		&board.ID,
		&board.OwnerID,
		&board.Name,
		&board.Description,
		&board.CreatedAt,
		&board.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Board{}, false
		}
		t.Fatalf("find board by id: %v", err)
	}

	return board, true
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

func GetColumnByID(t *testing.T, pool *pgxpool.Pool, columnID domain.ColumnID) (domain.Column, bool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const q = `
		SELECT id, board_id, name, description, position, created_at, updated_at
		FROM columns
		WHERE id = $1`

	var column domain.Column
	err := pool.QueryRow(ctx, q, columnID).Scan(
		&column.ID,
		&column.BoardID,
		&column.Name,
		&column.Description,
		&column.Position,
		&column.CreatedAt,
		&column.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Column{}, false
		}
		t.Fatalf("GetColumnByID() error = %v", err)
	}

	return column, true
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
		var column domain.Column
		err = rows.Scan(
			&column.ID,
			&column.BoardID,
			&column.Name,
			&column.Description,
			&column.Position,
			&column.CreatedAt,
			&column.UpdatedAt,
		)
		if err != nil {
			t.Fatalf("Column row Scan() error = %v", err)
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

func GetTaskByID(t *testing.T, pool *pgxpool.Pool, taskID domain.TaskID) (domain.Task, bool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const q = `
		SELECT id, column_id, name, description, position, created_at, updated_at
		FROM tasks
		WHERE id = $1`

	var task domain.Task
	err := pool.QueryRow(ctx, q, taskID).Scan(
		&task.ID,
		&task.ColumnID,
		&task.Name,
		&task.Description,
		&task.Position,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Task{}, false
		}
		t.Fatalf("GetTaskByID() error = %v", err)
	}

	return task, true
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
		var task domain.Task
		err = rows.Scan(
			&task.ID,
			&task.ColumnID,
			&task.Name,
			&task.Description,
			&task.Position,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			t.Fatalf("Task row Scan() error = %v", err)
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

func mustColumnPosition(t *testing.T, n int64) domain.ColumnPosition {
	t.Helper()

	p, err := domain.NewColumnPosition(n)
	if err != nil {
		t.Fatalf("NewColumnPosition(%d): %v", n, err)
	}

	return p
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

func assertTaskIDAndPosition(t *testing.T, task *domain.Task, wantID domain.TaskID, wantPos int64) {
	t.Helper()

	if task.ID != wantID {
		t.Errorf("got id %q, want %q", task.ID, wantID)
	}
	if task.Position.Int64() != wantPos {
		t.Errorf("got position %d, want %d", task.Position.Int64(), wantPos)
	}
}

func assertErrRowNotFound(t *testing.T, err error) {
	t.Helper()

	if !errors.Is(err, repository.ErrRowNotFound) {
		t.Errorf("got error %v, want ErrRowNotFound", err)
	}
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

func GetUserByID(t *testing.T, pool *pgxpool.Pool, userID domain.UserID) (domain.User, bool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const q = `SELECT id, email, password_hash, telegram_chat_id, telegram_username FROM users WHERE id = $1`

	var user domain.User
	err := pool.QueryRow(ctx, q, userID).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.TelegramChatID,
		&user.TelegramUsername,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, false
		}
		t.Fatalf("GetUserByID() error = %v", err)
	}

	return user, true
}

func setUserIDByTelegramLinkToken(t *testing.T, client *redis.Client, token domain.TelegramLinkToken, userID domain.UserID) {
	t.Helper()

	err := client.Set(context.Background(), telegramTokenPrefix+token.RevealSecret(), userID.String(), 0).Err()
	if err != nil {
		t.Fatalf("setTelegramTokenInRedis() error = %v", err)
	}
}

func getUserIDByTelegramLinkToken(t *testing.T, client *redis.Client, token domain.TelegramLinkToken) domain.UserID {
	t.Helper()

	val, err := client.Get(context.Background(), telegramTokenPrefix+token.RevealSecret()).Result()
	if err != nil {
		t.Fatalf("getTelegramTokenFromRedis() error = %v", err)
	}

	userID, err := domain.ParseUserID(val)
	if err != nil {
		t.Fatalf("getUserIDByTelegramLinkToken() error = %v", err)
	}

	return userID
}
