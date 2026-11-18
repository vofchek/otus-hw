package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

type Logger struct {
	slog      *slog.Logger
	slogLevel slog.Level
	logFile   *os.File
}

// создаёт объект логгера.
// внутри создаётся файл, в который будут складываться сообщения.
func New(level string, filePath string) (*Logger, error) {
	var logFile *os.File
	writer := io.Writer(os.Stdout)
	if filePath != "" {
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			return nil, fmt.Errorf("cant create log dir, err: %w", err)
		}

		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, fmt.Errorf("cant open log file, err: %w", err)
		}

		writer = io.MultiWriter(os.Stdout, f)
	}

	var slogLevel slog.Level
	if err := slogLevel.UnmarshalText([]byte(strings.TrimSpace(level))); err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", level, err)
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	return &Logger{
		slog:      slog.New(slog.NewJSONHandler(writer, opts)),
		slogLevel: slogLevel,
		logFile:   logFile,
	}, nil
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

func (l Logger) Close() error {
	return l.logFile.Close()
}
