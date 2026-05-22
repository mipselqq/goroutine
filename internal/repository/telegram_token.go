package repository

import (
	"context"
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
	err := r.redisClient.Set(
		ctx,
		telegramTokenPrefix+token.RevealSecret(),
		userID.String(),
		15*time.Minute,
	).Err()
	if err != nil {
		return fmt.Errorf("redis: insert link token: %v: %w", err, ErrInternal)
	}

	return nil
}
