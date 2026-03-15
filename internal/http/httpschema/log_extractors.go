package httpschema

import (
	"context"
	"log/slog"

	"goroutine/internal/domain"
	"goroutine/internal/logging"
)

func UserIDExtractor(ctx context.Context) []slog.Attr {
	if userID, ok := ctx.Value(ContextKeyUserID).(domain.UserID); ok {
		return []slog.Attr{slog.String("user_id", userID.String())}
	}
	return nil
}

func RequestIDExtractor(ctx context.Context) []slog.Attr {
	if reqID, ok := ctx.Value(ContextKeyRequestID).(string); ok {
		return []slog.Attr{slog.String("request_id", reqID)}
	}
	return nil
}

func AllExtractors() []logging.ContextExtractor {
	return []logging.ContextExtractor{
		UserIDExtractor,
		RequestIDExtractor,
	}
}
