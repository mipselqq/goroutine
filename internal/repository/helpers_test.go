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
)

func CreateUser(t *testing.T, pool *pgxpool.Pool, id domain.UserID, email string) {
	t.Helper()

	domainEmail, err := domain.NewEmail(email)
	if err != nil {
		t.Fatalf("create user email: %v", err)
	}

	InsertUser(t, pool, id, domainEmail, "hash")
}

func InsertUser(t *testing.T, pool *pgxpool.Pool, id domain.UserID, email domain.Email, hash string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const query = `INSERT INTO users (id, email, password_hash) VALUES ($1, $2, $3)`
	_, err := pool.Exec(ctx, query, id, email, hash)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}
}

func InsertBoard(t *testing.T, pool *pgxpool.Pool, board *domain.Board) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const q = `
		INSERT INTO boards (id, owner_id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := pool.Exec(ctx, q,
		board.ID,
		board.OwnerID,
		board.Name,
		board.Description,
		board.CreatedAt,
		board.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("InsertBoard() error = %v", err)
	}
}

func FindBoardByID(t *testing.T, pool *pgxpool.Pool, boardID domain.BoardID) (domain.Board, bool) {
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

func InsertColumn(t *testing.T, pool *pgxpool.Pool, column *domain.Column) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const q = `
		INSERT INTO columns (id, board_id, name, position, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := pool.Exec(ctx, q,
		column.ID,
		column.BoardID,
		column.Name,
		column.Position,
		column.CreatedAt,
		column.UpdatedAt,
	)
	if err != nil {
		t.Fatalf("InsertColumn() error = %v", err)
	}
}

func FindColumnByID(t *testing.T, pool *pgxpool.Pool, columnID domain.ColumnID) (domain.Column, bool) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const q = `
		SELECT id, board_id, name, position, created_at, updated_at
		FROM columns
		WHERE id = $1`

	var column domain.Column
	err := pool.QueryRow(ctx, q, columnID).Scan(
		&column.ID,
		&column.BoardID,
		&column.Name,
		&column.Position,
		&column.CreatedAt,
		&column.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Column{}, false
		}
		t.Fatalf("FindColumnByID() error = %v", err)
	}

	return column, true
}

func ListColumnsByBoardID(t *testing.T, pool *pgxpool.Pool, boardID domain.BoardID) []domain.Column {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const q = `
		SELECT id, board_id, name, position, created_at, updated_at
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
		err := rows.Scan(
			&column.ID,
			&column.BoardID,
			&column.Name,
			&column.Position,
			&column.CreatedAt,
			&column.UpdatedAt,
		)
		if err != nil {
			t.Fatalf("Column row Scan() error = %v", err)
		}

		columns = append(columns, column)
	}

	if err := rows.Err(); err != nil {
		t.Fatalf("rows.Err() error = %v", err)
	}

	return columns
}

func insertFixedUserAndBoard(t *testing.T, pool *pgxpool.Pool, email string) domain.Board {
	t.Helper()

	userID := testutil.ValidUserID()
	CreateUser(t, pool, userID, email)
	board := testutil.ValidBoard()
	InsertBoard(t, pool, &board)

	return board
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
