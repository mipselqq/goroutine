package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/driver"
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
		message, formatErr := TelegramMessageFromNotificationOutbox(ctx, &event)
		if formatErr != nil {
			return fmt.Errorf("create message: %v: %w", err, ErrInternal)
		}
		chatID, chatIDErr := domain.NewTelegramChatID(event.TelegramChatID)
		if chatIDErr != nil {
			return fmt.Errorf("create chat id: %v: %w", err, ErrInternal)
		}

		err = w.notifier.Notify(ctx, chatID, message)
		if err != nil {
			var responseErr *driver.ErrTelegramResponse
			if errors.Is(err, driver.ErrNetwork) {
				// Retry
				continue
			}
			if errors.As(err, &responseErr) {
				if slices.Contains([]int{429, 500, 502, 503, 504}, responseErr.Status) {
					// Retry
					continue
				}
				if responseErr.Status == 400 && responseErr.Type == "INPUT_USER_DEACTIVATED" {
					// Unlink Telegram
					continue
				}
				if slices.Contains([]int{400, 401, 403, 404}, responseErr.Status) {
					// Drop
					continue
				}
			}
			return fmt.Errorf("notify: %v: %w", err, ErrInternal)
		}
	}

	err = w.notificationRepo.Ack(ctx, ids)
	if err != nil {
		return fmt.Errorf("ack: %v: %w", err, ErrInternal)
	}

	return nil
}
