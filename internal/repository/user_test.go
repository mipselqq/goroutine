//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/testutil"
)

// TODO: factor out common initialization
func TestUserRepository_Insert(t *testing.T) {
	pool := testutil.SetupTestDB(t, "../../migrations")

	r := repository.NewPgUser(pool)
	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	emailStr := "test@example.com"
	email, _ := domain.NewEmail(emailStr)
	hash := "some-secret-hash"

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "users")

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
		testutil.TruncateTable(t, pool, "users")

		_ = r.Insert(ctx, email, hash)
		err := r.Insert(ctx, email, hash)

		if !errors.Is(err, repository.ErrUniqueViolation) {
			t.Errorf("Expected ErrUniqueViolation, got %v", err)
		}
	})
}

func TestUserRepository_GetPasswordHashByEmail(t *testing.T) {
	pool := testutil.SetupTestDB(t, "../../migrations")

	r := repository.NewPgUser(pool)

	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	emailStr := "test-get@example.com"
	email, _ := domain.NewEmail(emailStr)
	hash := "secret-hash-for-get"

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "users")

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
		testutil.TruncateTable(t, pool, "users")

		unknownEmail, _ := domain.NewEmail("unknown@example.com")
		_, err := r.GetPasswordHashByEmail(ctx, unknownEmail)

		if !errors.Is(err, repository.ErrRowNotFound) {
			t.Errorf("Expected ErrRowNotFound, got %v", err)
		}
	})
}
