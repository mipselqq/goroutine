// Package repository provides abstract data access for the application.
package repository

import (
	"context"
	"errors"
	"fmt"

	"goroutine/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PgUser struct {
	pool *pgxpool.Pool
}

func NewPgUser(pool *pgxpool.Pool) *PgUser {
	return &PgUser{
		pool: pool,
	}
}

const pgUniqueViolation = "23505"

func (r *PgUser) InsertUser(ctx context.Context, email domain.Email, hash string) error {
	const query = `INSERT INTO users (email, password_hash) VALUES ($1, $2)`

	_, err := r.pool.Exec(ctx, query, email, hash)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return fmt.Errorf("user repo: insert user: %w", ErrUniqueViolation)
	}

	return fmt.Errorf("user repo: insert user: %v: %w", err, ErrInternal)
}

func (r *PgUser) GetUserByEmail(ctx context.Context, email domain.Email) (id domain.UserID, hash string, err error) {
	const query = `SELECT id, password_hash FROM users WHERE email = $1`

	err = r.pool.QueryRow(ctx, query, email).Scan(&id, &hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.UserID{}, "", ErrRowNotFound
		}
		return domain.UserID{}, "", fmt.Errorf("user repo: get user by email: %v: %w", err, ErrInternal)
	}

	return id, hash, nil
}

func (r *PgUser) InsertTelegramLinkToken(ctx context.Context, token domain.TelegramLinkToken, userID domain.UserID) error {
	return fmt.Errorf("user repo: insert telegram link token: %w", ErrInternal)
}
