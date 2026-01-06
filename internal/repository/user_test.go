//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-todo/internal/domain"
	"go-todo/internal/repository"
	"go-todo/internal/testutil"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupUserRepository(t *testing.T) (*repository.PgUser, *pgxpool.Pool) {
	t.Helper()

	pool, _ := testutil.SetupTestDB(t)

	r := repository.NewPgUser(pool)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := pool.Exec(ctx, "TRUNCATE TABLE users CASCADE")
	if err != nil {
		t.Fatalf("Failed to TRUNCATE TABLE users: %v", err)
	}

	return r, pool
}

func TestUserRepository_Insert(t *testing.T) {
	r, pool := setupUserRepository(t)
	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	emailStr := "test@example.com"
	email, _ := domain.NewEmail(emailStr)
	hash := "some-secret-hash"

	t.Run("Success", func(t *testing.T) {
		err := r.Insert(ctx, email, hash)
		if err != nil {
			t.Errorf("Insert() error = %v", err)
		}

		var dbEmail string
		err = pool.QueryRow(ctx, "SELECT email FROM users WHERE email=$1", emailStr).Scan(&dbEmail)
		if err != nil {
			t.Errorf("Failed to find user in DB: %v", err)
		}
		if dbEmail != emailStr {
			t.Errorf("Expected email %s, got %s", emailStr, dbEmail)
		}
	})

	t.Run("Duplicate email", func(t *testing.T) {
		_, err := pool.Exec(ctx, "TRUNCATE TABLE users CASCADE")
		if err != nil {
			t.Fatalf("Failed to TRUNCATE TABLE users: %v", err)
		}
		_ = r.Insert(ctx, email, hash)
		err = r.Insert(ctx, email, hash)

		if !errors.Is(err, repository.ErrUniqueViolation) {
			t.Errorf("Expected ErrUniqueViolation, got %v", err)
		}
	})
}

func TestUserRepository_GetPasswordHashByEmail(t *testing.T) {
	r, pool := setupUserRepository(t)
	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	emailStr := "test-get@example.com"
	email, _ := domain.NewEmail(emailStr)
	hash := "secret-hash-for-get"

	t.Run("Success", func(t *testing.T) {
		err := r.Insert(ctx, email, hash)
		if err != nil {
			t.Fatalf("Failed to insert user for test: %v", err)
		}

		gotHash, err := r.GetPasswordHashByEmail(ctx, email)
		if err != nil {
			t.Errorf("GetPasswordHashByEmail() error = %v", err)
		}
		if gotHash != hash {
			t.Errorf("Expected hash %s, got %s", hash, gotHash)
		}
	})

	t.Run("User not found", func(t *testing.T) {
		unknownEmail, _ := domain.NewEmail("unknown@example.com")
		_, err := r.GetPasswordHashByEmail(ctx, unknownEmail)

		if !errors.Is(err, repository.ErrRowNotFound) {
			t.Errorf("Expected ErrRowNotFound, got %v", err)
		}
	})
}
