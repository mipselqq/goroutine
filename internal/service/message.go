package service

import (
	"context"
	"fmt"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
)

func (w *notificationWorker) TelegramMessageFromNotificationOutbox(ctx context.Context, event *repository.OutboxEvent) (domain.TelegramMessage, error) {
	notificationType, err := domain.NewNotificationType(event.EventType)
	if err != nil {
		return domain.TelegramMessage{}, fmt.Errorf("invalid notification type: %v", err)
	}

	payload, issues := ParseNotificationPayload(notificationType, event.Payload)
	if len(issues) > 0 {
		return domain.TelegramMessage{}, fmt.Errorf("invalid notification payload: %v", issues)
	}

	message := FormatNotificationMessage(domain.Notification{
		Type:    notificationType,
		Payload: payload,
	})
	telegramMessage, err := domain.NewTelegramMessage(message)
	if err != nil {
		return domain.TelegramMessage{}, fmt.Errorf("invalid telegram message: %v", err)
	}

	return telegramMessage, nil
}

func FormatNotificationMessage(notification domain.Notification) string {
	switch p := notification.Payload.(type) {
	case domain.BoardCreated:
		return fmt.Sprintf("%s created board %q", p.CallerEmail, p.BoardName)
	case domain.BoardUpdated:
		return fmt.Sprintf("%s updated board %q", p.CallerEmail, p.BoardName)
	case domain.BoardDeleted:
		return fmt.Sprintf("%s deleted board %q", p.CallerEmail, p.BoardName)
	case domain.ColumnCreated:
		return fmt.Sprintf("%s created column %q on board %q", p.CallerEmail, p.ColumnName, p.BoardName)
	case domain.ColumnUpdated:
		return fmt.Sprintf("%s updated column %q on board %q", p.CallerEmail, p.ColumnName, p.BoardName)
	case domain.ColumnMoved:
		return fmt.Sprintf("%s moved column %q on board %q from position %d to %d", p.CallerEmail, p.ColumnName, p.BoardName, p.SourcePosition.Int64(), p.TargetPosition.Int64())
	case domain.ColumnDeleted:
		return fmt.Sprintf("%s deleted column %q on board %q", p.CallerEmail, p.ColumnName, p.BoardName)
	case domain.TaskCreated:
		return fmt.Sprintf("%s created task %q in column %q on board %q", p.CallerEmail, p.TaskName, p.ColumnName, p.BoardName)
	case domain.TaskUpdated:
		return fmt.Sprintf("%s updated task %q in column %q on board %q", p.CallerEmail, p.TaskName, p.ColumnName, p.BoardName)
	case domain.TaskMoved:
		return fmt.Sprintf("%s moved task %q on board %q from %q (%d) to %q (%d)", p.CallerEmail, p.TaskName, p.BoardName, p.SourceColumnName, p.SourcePosition.Int64(), p.TargetColumnName, p.TargetPosition.Int64())
	case domain.TaskDeleted:
		return fmt.Sprintf("%s deleted task %q in column %q on board %q", p.CallerEmail, p.TaskName, p.ColumnName, p.BoardName)
	default:
		panic(fmt.Sprintf("BUG: unreachable message payload %T", notification.Payload))
	}
}
