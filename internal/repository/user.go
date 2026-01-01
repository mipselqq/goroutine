package repository

import (
	"context"
	"errors"

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

func (r *PgUser) Insert(ctx context.Context, email, hash string) error {
	const query = `INSERT INTO users (email, password_hash) VALUES ($1, $2)`

	_, err := r.pool.Exec(ctx, query, email, hash)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return ErrUniqueViolation
	}

	// TODO: wrap in a context
	return ErrInternal
}
