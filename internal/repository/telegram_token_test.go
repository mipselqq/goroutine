//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/testutil"

	"github.com/redis/go-redis/v9"
)

func TestRedisTelegramToken_InsertLinkToken(t *testing.T) {
	client, repo := telegramTokenRepoPrelude(t)

	ctx := context.Background()
	token := testutil.ValidTelegramLinkToken()
	userID := domain.NewUserID()

	err := repo.InsertLinkToken(ctx, token, userID)
	if err != nil {
		t.Fatalf("InsertLinkToken() error = %v, want nil", err)
	}

	value, err := client.Get(ctx, telegramTokenPrefix+token.RevealSecret()).Result()
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}
	if value != userID.String() {
		t.Fatalf("Get() value = %s, want %s", value, userID.String())
	}

	err = client.Del(ctx, telegramTokenPrefix+token.RevealSecret()).Err()
	if err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}
}

func TestRedisTelegramToken_InsertLinkToken_AlreadyExists(t *testing.T) {
	_, repo := telegramTokenRepoPrelude(t)

	ctx := context.Background()
	token := testutil.ValidTelegramLinkToken()
	userID := domain.NewUserID()

	err := repo.InsertLinkToken(ctx, token, userID)
	if err != nil {
		t.Fatalf("InsertLinkToken() first insert error = %v, want nil", err)
	}

	err = repo.InsertLinkToken(ctx, token, userID)
	if !errors.Is(err, repository.ErrTelegramLinkTokenAlreadyExists) {
		t.Fatalf("InsertLinkToken() second insert error = %v, want ErrTelegramLinkTokenAlreadyExists", err)
	}
}

func telegramTokenRepoPrelude(t *testing.T) (*redis.Client, *repository.RedisTelegramToken) {
	t.Helper()

	redisClient := testutil.SetupTestRedis(t)
	testutil.FlushCurrentRedisDB(t, redisClient)
	t.Cleanup(func() {
		err := redisClient.Conn().Close()
		if err != nil {
			t.Fatalf("Failed to close Redis connection")
		}
		testutil.FlushCurrentRedisDB(t, redisClient)
	})
	tokenRepo := repository.NewRedisTelegramToken(redisClient)

	return redisClient, tokenRepo
}
