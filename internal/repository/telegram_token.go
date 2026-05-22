package repository

import (
	"context"
	"errors"

	"goroutine/internal/domain"
)

type RedisTelegramToken struct{}

func NewRedisTelegramToken() *RedisTelegramToken {
	return &RedisTelegramToken{}
}

func (r *RedisTelegramToken) InsertLinkToken(ctx context.Context, token domain.TelegramLinkToken, userID domain.UserID) error {
	return errors.New("not implemented")
}
