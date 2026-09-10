package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"slices"
	"sync"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/driver"
	"goroutine/internal/logging"
	"goroutine/internal/repository"
)

const (
	actionRetry  = "retry"
	actionDrop   = "drop"
	actionUnlink = "unlink"
	actionExit   = "exit"

	minBackoff = 1 * time.Second
	maxBackoff = 5 * time.Minute
)

type notificationRepository interface {
	Claim(ctx context.Context, count int) ([]repository.OutboxEvent, error)
	Ack(ctx context.Context, ids []int64) error
	Retry(ctx context.Context, id int64, availableAt time.Time) error
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
		notifier:         notify,
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

	type sendResult struct {
		event *repository.OutboxEvent
		err   error
	}
	results := make([]sendResult, len(events))

	var wg sync.WaitGroup
	for i := range events {
		wg.Go(func() {
			results[i] = sendResult{event: &events[i], err: w.send(ctx, &events[i])}
		})
	}
	wg.Wait()

	var ack []int64
	for i := range results {
		result := &results[i]
		if result.err == nil {
			ack = append(ack, result.event.ID)
			continue
		}

		switch classify(result.err) {
		case actionRetry:
			err = w.notificationRepo.Retry(ctx, result.event.ID, retryAt(result.event, result.err))
			if err != nil {
				return fmt.Errorf("retry: %v: %w", err, ErrInternal)
			}
		case actionUnlink:
			w.logger.InfoContext(ctx, "dropping outbox event for deactivated telegram user", slog.Int64("id", result.event.ID))
			ack = append(ack, result.event.ID)
		case actionDrop:
			ack = append(ack, result.event.ID)
		case actionExit:
			if len(ack) > 0 {
				err = w.notificationRepo.Ack(ctx, ack)
				if err != nil {
					return fmt.Errorf("ack: %v: %w", err, ErrInternal)
				}
			}
			w.logger.ErrorContext(ctx, fmt.Sprintf("got unrecoverable status code %s, exiting...", unrecoverableStatus(result.err)))
			return result.err
		}
	}

	if len(ack) == 0 {
		return nil
	}

	err = w.notificationRepo.Ack(ctx, ack)
	if err != nil {
		return fmt.Errorf("ack: %v: %w", err, ErrInternal)
	}

	return nil
}

func (w *notificationWorker) send(ctx context.Context, event *repository.OutboxEvent) error {
	message, err := TelegramMessageFromNotificationOutbox(ctx, event)
	if err != nil {
		return fmt.Errorf("create message: %w", err)
	}

	chatID, err := domain.NewTelegramChatID(event.TelegramChatID)
	if err != nil {
		return fmt.Errorf("create chat id: %w", err)
	}

	return w.notifier.Notify(ctx, chatID, message)
}

func classify(err error) string {
	if errors.Is(err, driver.ErrNetwork) {
		return actionRetry
	}

	var tg *driver.ErrTelegramResponse
	if !errors.As(err, &tg) {
		return ""
	}

	if slices.Contains([]int{429, 500, 502, 503, 504}, tg.Status) {
		return actionRetry
	}
	if tg.Status == 400 && tg.Type == "INPUT_USER_DEACTIVATED" {
		return actionUnlink
	}
	if slices.Contains([]int{401, 404}, tg.Status) {
		return actionExit
	}
	if slices.Contains([]int{400, 403}, tg.Status) {
		return actionDrop
	}

	return ""
}

func unrecoverableStatus(err error) string {
	var tg *driver.ErrTelegramResponse
	if errors.As(err, &tg) {
		return http.StatusText(tg.Status)
	}
	return ""
}

func CalculateBackoff(attempts int, minDelay, maxDelay time.Duration) time.Duration {
	return min(maxDelay, time.Duration(float64(minDelay)*math.Pow(2, float64(attempts))))
}

func retryAt(event *repository.OutboxEvent, err error) time.Time {
	at := timeNow().Add(CalculateBackoff(event.Attempts, minBackoff, maxBackoff))
	var tg *driver.ErrTelegramResponse
	if errors.As(err, &tg) && tg.Status == 429 && tg.RetryAfter.After(at) {
		return tg.RetryAfter
	}
	return at
}
