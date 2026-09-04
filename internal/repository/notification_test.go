//go:build integration

package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v5/pgxpool"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/testutil"
)

func EventsForRecipients(userIDs ...domain.UserID) []repository.OutboxEvent {
	events := make([]repository.OutboxEvent, len(userIDs))
	for i, userID := range userIDs {
		events[i] = repository.OutboxEvent{
			RecipientUserID: userID.UUID(),
			EventType:       "board.created",
			Payload:         []byte(`{"boardName": "Work", "callerEmail": "a@x.com"}`),
		}
	}
	return events
}

func InsertOutboxEvents(t *testing.T, pool *pgxpool.Pool, events []repository.OutboxEvent) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	const query = `
		INSERT INTO notification_outbox (recipient_user_id, event_type, payload)
		VALUES ($1, $2, $3)`

	for _, event := range events {
		_, err := pool.Exec(ctx, query, event.RecipientUserID, event.EventType, event.Payload)
		if err != nil {
			t.Fatalf("InsertOutboxEvents() error = %v", err)
		}
	}
}

func TestNotificationRepository_Claim(t *testing.T) {
	pool, r := notificationRepoPrelude(t)

	testutil.TruncateAllTables(t, pool)

	userA := domain.NewUserID()
	userB := domain.NewUserID()
	userC := domain.NewUserID()
	emailC, err := domain.NewEmail("c@x.com")
	if err != nil {
		t.Fatalf("NewEmail() error = %v", err)
	}
	CreateUser(t, pool, userA, testutil.ValidEmail(), testutil.ValidPasswordHash())
	CreateUser(t, pool, userB, testutil.AnotherValidEmail(), testutil.ValidPasswordHash())
	CreateUser(t, pool, userC, emailC, testutil.ValidPasswordHash())

	events := EventsForRecipients(userA, userA, userB, userB, userC, userC)
	want := EventsForRecipients(userA, userB, userC, userA, userB)
	InsertOutboxEvents(t, pool, events)

	got, err := r.Claim(context.Background(), 5)
	if err != nil {
		t.Fatalf("Claim() error = %v", err)
	}

	diff := cmp.Diff(want, got, cmpopts.IgnoreFields(repository.OutboxEvent{}, "ID", "CreatedAt"))
	if diff != "" {
		t.Errorf("Claim() mismatch (-want +got):\n%s", diff)
	}
}

func TestNotificationRepository_Ack(t *testing.T) {
	pool, r := notificationRepoPrelude(t)

	testutil.TruncateAllTables(t, pool)

	user := domain.NewUserID()
	CreateUser(t, pool, user, testutil.ValidEmail(), testutil.ValidPasswordHash())
	InsertOutboxEvents(t, pool, EventsForRecipients(user, user))

	err := r.Ack(context.Background(), []int64{1, 2})
	if err != nil {
		t.Fatalf("Ack() error = %v", err)
	}

	AssertOutboxEvents(t, pool, []WantOutboxEvent{})
}

func notificationRepoPrelude(t *testing.T) (*pgxpool.Pool, *repository.PGNotification) {
	t.Helper()

	pool := testutil.SetupPostgres(t)
	t.Cleanup(func() { pool.Close() })

	return pool, repository.NewPGNotification(pool)
}
