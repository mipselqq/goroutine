package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"goroutine/internal/domain"

	"github.com/redis/go-redis/v9"
)

type RedisTelegramToken struct {
	redisClient *redis.Client
}

func NewRedisTelegramToken(redisClient *redis.Client) *RedisTelegramToken {
	return &RedisTelegramToken{redisClient: redisClient}
}

const telegramTokenPrefix = "tg_token:"

func (r *RedisTelegramToken) InsertLinkToken(ctx context.Context, token domain.TelegramLinkToken, userID domain.UserID) error {
	inserted, err := r.redisClient.SetNX(
		ctx,
		telegramTokenPrefix+token.RevealSecret(),
		userID.String(),
		15*time.Minute,
	).Result()
	if err != nil {
		return fmt.Errorf("redis: insert link token: %v: %w", err, ErrInternal)
	}

	if !inserted {
		return fmt.Errorf("redis: insert link token: %w", ErrTelegramLinkTokenAlreadyExists)
	}

	return nil
}

func (r *RedisTelegramToken) GetUserIDByLinkToken(ctx context.Context, token domain.TelegramLinkToken) (domain.UserID, error) {
	userIDStr, err := r.redisClient.Get(
		ctx,
		telegramTokenPrefix+token.RevealSecret(),
	).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return domain.UserID{}, fmt.Errorf("redis: get user id by link token: token not found: %w", ErrTelegramLinkTokenNotFound)
		}
		return domain.UserID{}, fmt.Errorf("redis: get user id by link token: %v: %w", err, ErrInternal)
	}

	userID, err := domain.ParseUserID(userIDStr)
	if err != nil {
		return domain.UserID{}, fmt.Errorf("redis: get user id by link token: parse user id: corrupted data: %v: %w", err, ErrInternal)
	}

	return userID, nil
}
