package logging

import (
	"fmt"
	"log/slog"
	"os"
)

type GooseLogger struct {
	Base *slog.Logger
}

func (l *GooseLogger) Fatal(v ...any) {
	l.Base.Error(fmt.Sprint(v...))
	os.Exit(1)
}

func (l *GooseLogger) Fatalf(format string, v ...any) {
	l.Base.Error(fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (l *GooseLogger) Print(v ...any) {
	l.Base.Info(fmt.Sprint(v...))
}

func (l *GooseLogger) Printf(format string, v ...any) {
	l.Base.Info(fmt.Sprintf(format, v...))
}
