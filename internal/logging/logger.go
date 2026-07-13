package logging

import (
	"context"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

type ContextExtractor func(ctx context.Context) []slog.Attr

type ContextHandler struct {
	Base       slog.Handler
	extractors []ContextExtractor
}

func (h *ContextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Base.Enabled(ctx, level)
}

//nolint:gocritic // r must be passed by value to implement slog.Handler
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, extractor := range h.extractors {
		if attrs := extractor(ctx); len(attrs) > 0 {
			r.AddAttrs(attrs...)
		}
	}
	return h.Base.Handle(ctx, r)
}

func (h *ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &ContextHandler{
		Base:       h.Base.WithAttrs(attrs),
		extractors: h.extractors,
	}
}

func (h *ContextHandler) WithGroup(name string) slog.Handler {
	return &ContextHandler{
		Base:       h.Base.WithGroup(name),
		extractors: h.extractors,
	}
}

func NewLogger(env, levelStr string, extractors ...ContextExtractor) *slog.Logger {
	var handler slog.Handler
	level := parseLevel(levelStr)

	if env == "dev" {
		handler = tint.NewTextHandler(os.Stdout, &tint.Options{Level: level})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	if len(extractors) > 0 {
		handler = &ContextHandler{
			Base:       handler,
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
