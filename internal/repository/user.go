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

func (r *PgUser) GetUserByEmail(ctx context.Context, email domain.Email) (domain.User, error) {
	const query = `SELECT id, email, password_hash, telegram_chat_id, telegram_username FROM users WHERE email = $1`

	var user domain.User
	err := r.pool.QueryRow(ctx, query, email).Scan( // TODO: Maybe implement Scan for the struct itself?
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.TelegramChatID,
		&user.TelegramUsername,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrRowNotFound
		}
		return domain.User{}, fmt.Errorf("user repo: get user by email: %v: %w", err, ErrInternal)
	}

	return user, nil
}

func (r *PgUser) UpdateTelegramInfo(ctx context.Context, userID domain.UserID, chatID domain.TelegramChatID, username domain.TelegramUsername) error {
	const query = `UPDATE users SET telegram_chat_id = $1, telegram_username = $2 WHERE id = $3`

	_, err := r.pool.Exec(ctx, query, chatID, username, userID)
	if err != nil {
		return fmt.Errorf("user repo: update telegram info: %v: %w", err, ErrInternal)
	}

	return nil
}
