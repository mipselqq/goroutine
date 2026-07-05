package testutil

import (
	"context"
	"testing"

	"goroutine/internal/app"

	"github.com/redis/go-redis/v9"
)

func SetupRedis(t *testing.T) *redis.Client {
	t.Helper()
	MustLoadDevEnv()
	logger := NewLogger(t)

	client, err := app.SetupRedisFromEnv(logger)
	if err != nil {
		t.Fatalf("SetupRedisFromEnv() error = %v", err)
	}

	return client
}

func FlushRedisDB(t *testing.T, client *redis.Client) {
	t.Helper()

	err := client.FlushDB(context.Background()).Err()
	if err != nil {
		t.Fatalf("FlushDB() error = %v", err)
	}
}
