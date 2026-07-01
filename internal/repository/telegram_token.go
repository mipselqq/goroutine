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
	redisClient  *redis.Client
	linkTokenTTL time.Duration
}

func NewRedisTelegramToken(redisClient *redis.Client, linkTokenTTL time.Duration) *RedisTelegramToken {
	return &RedisTelegramToken{
		redisClient:  redisClient,
		linkTokenTTL: linkTokenTTL,
	}
}

const telegramTokenPrefix = "tg_token:"

func tokenToKey(token domain.TelegramLinkToken) string {
	return telegramTokenPrefix + token.RevealSecret()
}

func (r *RedisTelegramToken) InsertLinkToken(ctx context.Context, token domain.TelegramLinkToken, userID domain.UserID) error {
	inserted, err := r.redisClient.SetNX(
		ctx,
		tokenToKey(token),
		userID.String(),
		r.linkTokenTTL,
	).Result()
	if err != nil {
		return fmt.Errorf("redis: insert link token: %v: %w", err, ErrInternal)
	}

	if !inserted {
		return fmt.Errorf("redis: insert link token: %w", ErrKeyExists)
	}

	return nil
}

func (r *RedisTelegramToken) ConsumeTelegramLinkToken(ctx context.Context, token domain.TelegramLinkToken) (domain.UserID, error) {
	userIDStr, err := r.redisClient.GetDel(ctx, tokenToKey(token)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return domain.UserID{}, fmt.Errorf("redis: consume telegram link token: token not found: %w", ErrKeyNotFound)
		}
		return domain.UserID{}, fmt.Errorf("redis: consume telegram link token: %v: %w", err, ErrInternal)
	}

	userID, err := domain.ParseUserID(userIDStr)
	if err != nil {
		return domain.UserID{}, fmt.Errorf("redis: consume telegram link token: parse user id: corrupted data: %v: %w", err, ErrInternal)
	}

	return userID, nil
}
