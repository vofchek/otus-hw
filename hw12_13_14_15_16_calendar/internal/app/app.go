package app

import (
	"context"
	"errors"

	"github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage"
)

type App struct {
	logger  Logger
	storage Storage
}

type Logger interface {
	Info(msg string)
	Error(msg string)
	Log(ctx context.Context, msg string, args ...any)
}

type Storage interface {
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
	Add(ctx context.Context, event storage.Event) error
	Update(ctx context.Context, event storage.Event) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*storage.Event, error)
	GetAll(ctx context.Context) ([]storage.Event, error)
}

func New(logger Logger, storage Storage) *App {
	return &App{
		logger,
		storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, title string) error {
	_ = ctx
	_ = title
	return errors.New("not implemented")
}

// TODO
