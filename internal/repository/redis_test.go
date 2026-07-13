package repository_test

import (
	"context"
	"testing"

	"goroutine/internal/domain"

	"github.com/redis/go-redis/v9"
)

const telegramTokenPrefix = "tg_token:"

func setUserIDByTelegramLinkToken(t *testing.T, client *redis.Client, token domain.TelegramLinkToken, userID domain.UserID) {
	t.Helper()

	err := client.Set(context.Background(), telegramTokenPrefix+token.RevealSecret(), userID.String(), 0).Err()
	if err != nil {
		t.Fatalf("setTelegramTokenInRedis() error = %v", err)
	}
}

func getUserIDByTelegramLinkToken(t *testing.T, client *redis.Client, token domain.TelegramLinkToken) domain.UserID {
	t.Helper()

	val, err := client.Get(context.Background(), telegramTokenPrefix+token.RevealSecret()).Result()
	if err != nil {
		t.Fatalf("getTelegramTokenFromRedis() error = %v", err)
	}

	userID, err := domain.ParseUserID(val)
	if err != nil {
		t.Fatalf("getUserIDByTelegramLinkToken() error = %v", err)
	}

	return userID
}
