package logging

import (
	"fmt"
	"log/slog"
	"os"
)

type GooseLogger struct {
	Logger *slog.Logger
}

func (l *GooseLogger) Fatal(v ...interface{}) {
	l.Logger.Error(fmt.Sprint(v...))
	os.Exit(1)
}

func (l *GooseLogger) Fatalf(format string, v ...interface{}) {
	l.Logger.Error(fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (l *GooseLogger) Print(v ...interface{}) {
	l.Logger.Info(fmt.Sprint(v...))
}

func (l *GooseLogger) Printf(format string, v ...interface{}) {
	l.Logger.Info(fmt.Sprintf(format, v...))
}
