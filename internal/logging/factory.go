package logging

import (
	"log/slog"
)

func NewLoggerContext(logger *slog.Logger, module string) *slog.Logger {
	return logger.With(slog.String("module", module))
}

func NewDiscardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}
