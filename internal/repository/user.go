// Package repository provides abstract data access for the application.
package repository

import (
	"context"
	"errors"
	"fmt"

	"goroutine/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGUser struct {
	pgPool *pgxpool.Pool
}

func NewPGUser(pgPool *pgxpool.Pool) *PGUser {
	return &PGUser{
		pgPool: pgPool,
	}
}

const pgUniqueViolation = "23505"

func (r *PGUser) Create(ctx context.Context, email domain.Email, hash string) error {
	const query = `INSERT INTO users (email, password_hash) VALUES ($1, $2)`

	_, err := r.pgPool.Exec(ctx, query, email, hash)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return fmt.Errorf("user repo: insert user: %w", ErrUniqueViolation)
	}

	return fmt.Errorf("user repo: insert user: %v: %w", err, ErrInternal)
}

func (r *PGUser) GetByEmail(ctx context.Context, email domain.Email) (domain.User, error) {
	const query = `SELECT id, email, password_hash, telegram_chat_id, telegram_username FROM users WHERE email = $1`

	user, err := ScanUser(r.pgPool.QueryRow(ctx, query, email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrRowNotFound
		}
		return domain.User{}, fmt.Errorf("user repo: get user by email: %v: %w", err, ErrInternal)
	}

	return user, nil
}

func (r *PGUser) UpdateTelegramInfo(ctx context.Context, userID domain.UserID, chatID domain.TelegramChatID, username domain.TelegramUsername) error {
	const query = `UPDATE users SET telegram_chat_id = $1, telegram_username = $2 WHERE id = $3`

	status, err := r.pgPool.Exec(ctx, query, chatID, username, userID)
	if err != nil {
		return fmt.Errorf("user repo: update telegram info: %v: %w", err, ErrInternal)
	}

	if status.RowsAffected() == 0 {
		return fmt.Errorf("user repo: update telegram info: %w", ErrRowNotFound)
	}

	return nil
}

func ScanUser(row interface{ Scan(...any) error }) (domain.User, error) {
	var (
		rawID               uuid.UUID
		rawEmail            string
		rawPasswordHash     string
		rawTelegramChatID   *int64
		rawTelegramUsername *string
	)
	err := row.Scan(&rawID, &rawEmail, &rawPasswordHash, &rawTelegramChatID, &rawTelegramUsername)
	if err != nil {
		return domain.User{}, fmt.Errorf("scan user: %w", err)
	}

	email, err := domain.NewEmail(rawEmail)
	if err != nil {
		return domain.User{}, fmt.Errorf("scan user: email: %w: %w", domain.ErrDataCorrupted, err)
	}

	var chatID domain.TelegramChatID
	if rawTelegramChatID != nil {
		chatID, err = domain.NewTelegramChatID(*rawTelegramChatID)
		if err != nil {
			return domain.User{}, fmt.Errorf("scan user: telegram chat id: %w: %w", domain.ErrDataCorrupted, err)
		}
	}

	var username domain.TelegramUsername
	if rawTelegramUsername != nil {
		username, err = domain.NewTelegramUsername(*rawTelegramUsername)
		if err != nil {
			return domain.User{}, fmt.Errorf("scan user: telegram username: %w: %w", domain.ErrDataCorrupted, err)
		}
	}

	return domain.User{
		ID:               domain.UserIDFromUUID(rawID),
		Email:            email,
		PasswordHash:     rawPasswordHash,
		TelegramChatID:   chatID,
		TelegramUsername: username,
	}, nil
}
