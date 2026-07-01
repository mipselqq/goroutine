//go:build integration

package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/testutil"

	"github.com/redis/go-redis/v9"
)

func TestTelegramTokenRepository_InsertLinkToken(t *testing.T) {
	client, repo := telegramTokenRepoPrelude(t)

	ctx := context.Background()
	token := testutil.ValidTelegramLinkToken()
	userID := domain.NewUserID()

	err := repo.InsertLinkToken(ctx, token, userID)
	if err != nil {
		t.Fatalf("InsertLinkToken() error = %v, want nil", err)
	}

	redisUserID := getUserIDByTelegramLinkToken(t, client, token)
	if redisUserID != userID {
		t.Fatalf("Get() value = %s, want %s", redisUserID, userID.String())
	}
}

func TestTelegramTokenRepository_InsertLinkToken_AlreadyExists(t *testing.T) {
	client, repo := telegramTokenRepoPrelude(t)

	ctx := context.Background()
	token := testutil.ValidTelegramLinkToken()
	userID := domain.NewUserID()
	anotherUserID := domain.NewUserID()

	err := repo.InsertLinkToken(ctx, token, userID)
	if err != nil {
		t.Fatalf("InsertLinkToken() first insert error = %v, want nil", err)
	}

	err = repo.InsertLinkToken(ctx, token, anotherUserID)
	if !errors.Is(err, repository.ErrTelegramLinkTokenAlreadyExists) {
		t.Fatalf("InsertLinkToken() second insert error = %v, want ErrTelegramLinkTokenAlreadyExists", err)
	}

	redisUserID := getUserIDByTelegramLinkToken(t, client, token)
	if redisUserID != userID {
		t.Fatalf("ConsumeTelegramLinkToken() userID = %s, want %s", redisUserID, userID.String())
	}
}

func TestTelegramTokenRepository_ConsumeTelegramLinkToken(t *testing.T) {
	client, repo := telegramTokenRepoPrelude(t)

	ctx := context.Background()
	token := testutil.ValidTelegramLinkToken()
	userID := domain.NewUserID()

	_, err := repo.ConsumeTelegramLinkToken(ctx, token)
	if !errors.Is(err, repository.ErrTelegramLinkTokenNotFound) {
		t.Fatalf("ConsumeTelegramLinkToken() error = %v, want ErrTelegramLinkTokenNotFound", err)
	}

	setUserIDByTelegramLinkToken(t, client, token, userID)

	userIDFromToken, err := repo.ConsumeTelegramLinkToken(ctx, token)
	if err != nil {
		t.Fatalf("ConsumeTelegramLinkToken() error = %v, want nil", err)
	}
	if userIDFromToken != userID {
		t.Fatalf("ConsumeTelegramLinkToken() userID = %v, want %v", userIDFromToken, userID)
	}

	_, err = repo.ConsumeTelegramLinkToken(ctx, token)
	if !errors.Is(err, repository.ErrTelegramLinkTokenNotFound) {
		t.Fatalf("ConsumeTelegramLinkToken() second call error = %v, want ErrTelegramLinkTokenNotFound", err)
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
	tokenRepo := repository.NewRedisTelegramToken(redisClient, 15*time.Minute)

	return redisClient, tokenRepo
}
