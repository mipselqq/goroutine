//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/testutil"
)

func TestUserRepository_Insert(t *testing.T) {
	pool, r := userRepoPrelude(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	email := testutil.ValidEmail()
	hash := testutil.ValidPasswordHash()

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		err := r.Insert(ctx, email, hash)
		if err != nil {
			t.Errorf("Insert() error = %v", err)
		}

		var dbEmail domain.Email
		err = pool.QueryRow(ctx, "SELECT email FROM users WHERE email=$1", email).Scan(&dbEmail)
		if err != nil {
			t.Errorf("User row Scan() error = %v", err)
		}
		if dbEmail != email {
			t.Errorf("got email %q, want %q", dbEmail, email)
		}
	})

	t.Run("Duplicate email", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		InsertUser(t, pool, testutil.ValidUserID(), email, hash)
		err := r.Insert(ctx, email, hash)

		if !errors.Is(err, repository.ErrUniqueViolation) {
			t.Errorf("got error %v, want ErrUniqueViolation", err)
		}
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	pool, r := userRepoPrelude(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	email := testutil.ValidEmail()
	hash := testutil.ValidPasswordHash()
	userID := testutil.ValidUserID()

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		InsertUser(t, pool, userID, email, hash)

		gotID, gotHash, err := r.GetByEmail(ctx, email)
		if err != nil {
			t.Errorf("GetByEmail() error = %v", err)
		}
		if gotHash != hash {
			t.Errorf("got hash %q, want %q", gotHash, hash)
		}
		if gotID != userID {
			t.Errorf("got id %q, want %q", gotID, userID)
		}
	})

	t.Run("User not found", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		unknownEmail, _ := domain.NewEmail("unknown@example.com")
		_, _, err := r.GetByEmail(ctx, unknownEmail)

		assertErrRowNotFound(t, err)
	})
}

func userRepoPrelude(t *testing.T) (*pgxpool.Pool, *repository.PgUser) {
	t.Helper()

	pool := testutil.SetupTestDB(t, "../../migrations")
	t.Cleanup(func() { pool.Close() })

	return pool, repository.NewPgUser(pool)
}
