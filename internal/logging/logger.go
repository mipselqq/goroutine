package logging

import (
	"context"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

type ContextExtractor func(ctx context.Context) []slog.Attr

type ContextHandler struct {
	slog.Handler
	extractors []ContextExtractor
}

//nolint:gocritic // r must be passed by value to implement slog.Handler
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, extractor := range h.extractors {
		if attrs := extractor(ctx); len(attrs) > 0 {
			r.AddAttrs(attrs...)
		}
	}
	return h.Handler.Handle(ctx, r)
}

func NewLogger(env, logLevel string, extractors ...ContextExtractor) *slog.Logger {
	var handler slog.Handler
	level := parseLevel(logLevel)

	if env == "dev" {
		handler = tint.NewHandler(os.Stdout, &tint.Options{Level: level})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	if len(extractors) > 0 {
		handler = &ContextHandler{
			Handler:    handler,
			extractors: extractors,
		}
	}

	return slog.New(handler)
}

func WithModule(l *slog.Logger, module string) *slog.Logger {
	return l.With(slog.String("module", module))
}

func parseLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
