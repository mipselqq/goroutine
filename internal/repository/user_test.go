//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-todo/internal/config"
	"go-todo/internal/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func TestUserRepository_Insert(t *testing.T) {
	// Arrange
	// TODO: factor out to TestMain
	_ = godotenv.Load("../../.env")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	config, err := config.NewPgConfigFromEnv().ParsePgxpoolConfig()
	if err != nil {
		t.Fatalf("Failed to parse db config from env: %v", err)
	}

	// TODO: connect once inside TestMain
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatalf("Failed to connect to test db: %v", err)
	}
	defer pool.Close()

	r := repository.NewPgUser(pool)

	// TODO: run migrations outside?
	_, err = pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatalf("Failed to create table users: %v", err)
	}

	_, err = pool.Exec(ctx, "TRUNCATE TABLE users CASCADE")
	if err != nil {
		t.Fatalf("Failed to TRUNCATE TABLE users: %v", err)
	}

	email := "test@example.com"
	hash := "some-secret-hash"

	t.Run("Success", func(t *testing.T) {
		// Act
		err := r.Insert(ctx, email, hash)
		if err != nil {
			t.Errorf("Insert() error = %v", err)
		}

		var dbEmail string
		err = pool.QueryRow(ctx, "SELECT email FROM users WHERE email=$1", email).Scan(&dbEmail)
		if err != nil {
			t.Errorf("Failed to find user in DB: %v", err)
		}
		if dbEmail != email {
			t.Errorf("Expected email %s, got %s", email, dbEmail)
		}
	})

	t.Run("Duplicate email", func(t *testing.T) {
		// TODO: truncate as well
		_ = r.Insert(ctx, email, hash)
		err := r.Insert(ctx, email, hash)

		if !errors.Is(err, repository.ErrUniqueViolation) {
			t.Errorf("Expected ErrUniqueViolation, got %v", err)
		}
	})
}
