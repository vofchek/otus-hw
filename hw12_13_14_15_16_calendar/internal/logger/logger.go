package logger

import (
	"context"
	"io"
	"log"
	"log/slog"
	"strings"
)

type Logger struct {
	slog      *slog.Logger
	slogLevel slog.Level
}

func New(level string, w io.Writer) *Logger {
	var slogLevel slog.Level
	if err := slogLevel.UnmarshalText([]byte(strings.TrimSpace(level))); err != nil {
		log.Fatalf("invalid log level %q: %v", level, err)
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	return &Logger{
		slog:      slog.New(slog.NewTextHandler(w, opts)),
		slogLevel: slogLevel,
	}
}

func (l Logger) Info(msg string) {
	l.slog.Info(msg)
}

func (l Logger) Error(msg string) {
	l.slog.Error(msg)
}

func (l Logger) Log(ctx context.Context, msg string, args ...any) {
	l.slog.Log(ctx, l.slogLevel, msg, args...)
}
