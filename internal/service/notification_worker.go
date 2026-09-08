package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/logging"
	"goroutine/internal/repository"
)

type notificationRepository interface {
	Claim(ctx context.Context, count int) ([]repository.OutboxEvent, error)
	Ack(ctx context.Context, ids []int64) error
}

type Notifier interface {
	Notify(ctx context.Context, chatID domain.TelegramChatID, text domain.TelegramMessage) error
}

type notificationWorker struct {
	notificationRepo notificationRepository
	notifier         Notifier
	logger           *slog.Logger
	pollInterval     time.Duration
	claimBatchSize   int
}

func NewNotificationWorker(
	logger *slog.Logger,
	notificationRepo notificationRepository,
	notify Notifier,
	pollInterval time.Duration,
	claimBatchSize int,
) *notificationWorker {
	return &notificationWorker{
		notificationRepo: notificationRepo,
		logger:           logging.WithModule(logger, "service.notification_worker"),
		pollInterval:     pollInterval,
		claimBatchSize:   claimBatchSize,
	}
}

func (w *notificationWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			err := w.processBatch(ctx)
			if err != nil {
				return err
			}
		}
	}
}

func (w *notificationWorker) processBatch(ctx context.Context) error {
	events, err := w.notificationRepo.Claim(ctx, w.claimBatchSize)
	if err != nil {
		return fmt.Errorf("claim: %v: %w", err, ErrInternal)
	}
	if len(events) == 0 {
		return nil
	}

	ids := make([]int64, len(events))
	for i, event := range events {
		ids[i] = event.ID
		message, formatErr := w.TelegramMessageFromNotificationOutbox(ctx, &event)
		if formatErr != nil {
			return fmt.Errorf("create message: %v: %w", err, ErrInternal)
		}

		err = w.notifier.Notify(ctx, domain.TelegramChatID{}, message)
		if err != nil {
			return fmt.Errorf("notify: %v: %w", err, ErrInternal)
		}
	}

	err = w.notificationRepo.Ack(ctx, ids)
	if err != nil {
		return fmt.Errorf("ack: %v: %w", err, ErrInternal)
	}

	return nil
}
