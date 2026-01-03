package logging

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
)

func NewLoggerContext(logger *slog.Logger, module string) *slog.Logger {
	return logger.With(slog.String("module", module))
}

func NewDiscardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
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

func NewHandler(env, logLevel string) slog.Handler {
	if env == "dev" {
		return tint.NewHandler(os.Stdout, &tint.Options{
			Level: parseLevel(logLevel),
		})
	}
	return slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: parseLevel(logLevel),
	})
}

func NewLogger(env, logLevel string) *slog.Logger {
	return slog.New(NewHandler(env, logLevel))
}
