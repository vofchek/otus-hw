package logger

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

var ErrInvalidLogLevel = errors.New("invalid log level")

const (
	infoLevel  = "info"
	errorLevel = "error"
)

type Logger struct {
	slog      *slog.Logger
	slogLevel slog.Level
	logFile   *os.File
}

// создаёт объект логгера.
// внутри создаётся файл, в который будут складываться сообщения.
func New(level string, filePath string) (*Logger, error) {
	var slogLevel slog.Level
	if err := slogLevel.UnmarshalText([]byte(strings.TrimSpace(level))); err != nil {
		return nil, fmt.Errorf("%w %q: %w", ErrInvalidLogLevel, level, err)
	}

	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	var logFile *os.File
	writer := io.Writer(os.Stdout)
	if filePath != "" {
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			return nil, fmt.Errorf("cant create log dir, err: %w", err)
		}

		var err error
		logFile, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return nil, fmt.Errorf("cant open log file, err: %w", err)
		}

		writer = io.MultiWriter(os.Stdout, logFile)
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

// общая функция логгирования.
// level - строка с названием уровня логгирования, для уровней созданы констансты [infoLevel], [errorLevel].
func (l Logger) Log(ctx context.Context, level string, msg string, args ...any) error {
	var slogLevel slog.Level
	if err := slogLevel.UnmarshalText([]byte(strings.TrimSpace(string(level)))); err != nil {
		return fmt.Errorf("%w %q: %w", ErrInvalidLogLevel, level, err)
	}

	l.slog.Log(ctx, slogLevel, msg, args...)

	return nil
}

func (l Logger) Close() error {
	if l.logFile == nil {
		return nil
	}
	return l.logFile.Close()
}
