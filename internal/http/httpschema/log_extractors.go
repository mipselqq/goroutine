package httpschema

import (
	"context"
	"log/slog"

	"goroutine/internal/domain"
	"goroutine/internal/logging"
)

func userIDExtractor(ctx context.Context) []slog.Attr {
	if userID, ok := ctx.Value(ContextKeyUserID).(domain.UserID); ok {
		return []slog.Attr{slog.Any("user_id", userID)}
	}
	return nil
}

func requestIDExtractor(ctx context.Context) []slog.Attr {
	if reqID, ok := ctx.Value(ContextKeyRequestID).(string); ok {
		return []slog.Attr{slog.String("request_id", reqID)}
	}
	return nil
}

func AllExtractors() []logging.ContextExtractor {
	return []logging.ContextExtractor{
		userIDExtractor,
		requestIDExtractor,
	}
}
