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

func TestUserRepository_InsertUser(t *testing.T) {
	pool, r := userRepoPrelude(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	email := testutil.ValidEmail()
	hash := testutil.ValidPasswordHash()

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		err := r.InsertUser(ctx, email, hash)
		if err != nil {
			t.Errorf("InsertUser() error = %v", err)
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
		err := r.InsertUser(ctx, email, hash)

		if !errors.Is(err, repository.ErrUniqueViolation) {
			t.Errorf("got error %v, want ErrUniqueViolation", err)
		}
	})
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	pool, r := userRepoPrelude(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	email := testutil.ValidEmail()
	hash := testutil.ValidPasswordHash()
	userID := testutil.ValidUserID()

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		InsertUser(t, pool, userID, email, hash)

		user, err := r.GetUserByEmail(ctx, email)
		if err != nil {
			t.Fatalf("GetUserByEmail() error = %v", err)
		}
		if user.PasswordHash != hash {
			t.Errorf("got hash %q, want %q", user.PasswordHash, hash)
		}
		if user.ID != userID {
			t.Errorf("got id %q, want %q", user.ID, userID)
		}
	})

	t.Run("User not found", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		unknownEmail, _ := domain.NewEmail("unknown@example.com")
		_, err := r.GetUserByEmail(ctx, unknownEmail)

		assertErrRowNotFound(t, err)
	})
}

func TestUserRepository_UpdateTelegramInfo(t *testing.T) {
	pool, r := userRepoPrelude(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userID := testutil.ValidUserID()
	chatID := testutil.ValidTelegramChatID()
	username := testutil.ValidTelegramUsername()

	t.Run("Success", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		InsertUser(t, pool, userID, testutil.ValidEmail(), testutil.ValidPasswordHash())
		err := r.UpdateTelegramInfo(ctx, userID, chatID, username)
		if err != nil {
			t.Fatalf("UpdateTelegramInfo() error = %v", err)
		}

		user, found := FindUserByID(t, pool, userID)
		if !found {
			t.Fatal("FindUserByID() user not found")
		}
		if user.TelegramChatID != chatID {
			t.Errorf("got chatID %v, want %v", user.TelegramChatID, chatID)
		}
		if user.TelegramUsername != username {
			t.Errorf("got username %v, want %v", user.TelegramUsername, username)
		}
	})

	t.Run("User not found", func(t *testing.T) {
		testutil.TruncateAllTables(t, pool)

		err := r.UpdateTelegramInfo(ctx, userID, chatID, username)
		assertErrRowNotFound(t, err)
	})
}

func userRepoPrelude(t *testing.T) (*pgxpool.Pool, *repository.PgUser) {
	t.Helper()

	pool := testutil.SetupTestPostgres(t, "../../migrations")
	t.Cleanup(func() { pool.Close() })

	return pool, repository.NewPgUser(pool)
}
