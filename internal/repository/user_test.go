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

func TestUserRepository_Insert(t *testing.T) {
	pool := testutil.SetupTestDB(t, "../../migrations")

	r := repository.NewPgUser(pool)
	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	email := testutil.ValidEmail()
	hash := testutil.ValidPasswordHash()

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "users")

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
		testutil.TruncateTable(t, pool, "users")

		_ = r.Insert(ctx, email, hash)
		err := r.Insert(ctx, email, hash)

		if !errors.Is(err, repository.ErrUniqueViolation) {
			t.Errorf("got error %v, want ErrUniqueViolation", err)
		}
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	pool := testutil.SetupTestDB(t, "../../migrations")

	r := repository.NewPgUser(pool)

	defer pool.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	email := testutil.ValidEmail()
	hash := testutil.ValidPasswordHash()

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "users")

		err := r.Insert(ctx, email, hash)
		if err != nil {
			t.Fatalf("Insert() error = %v", err)
		}

		gotID, gotHash, err := r.GetByEmail(ctx, email)
		if err != nil {
			t.Errorf("GetByEmail() error = %v", err)
		}
		if gotHash != hash {
			t.Errorf("got hash %q, want %q", gotHash, hash)
		}
		if gotID.IsEmpty() {
			t.Errorf("got id %v, want non-empty", gotID)
		}
	})

	t.Run("User not found", func(t *testing.T) {
		testutil.TruncateTable(t, pool, "users")

		unknownEmail, _ := domain.NewEmail("unknown@example.com")
		_, _, err := r.GetByEmail(ctx, unknownEmail)

		if !errors.Is(err, repository.ErrRowNotFound) {
			t.Errorf("got error %v, want ErrRowNotFound", err)
		}
	})
}
