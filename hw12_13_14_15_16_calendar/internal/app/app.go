package app

import (
	"context"

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
	Add(event *storage.Event) error
	Update(event storage.Event) error
	Delete(id string) error
	GetById(id string) (*storage.Event, error)
	GetAll() ([]*storage.Event, error)
}

func New(logger Logger, storage Storage) *App {
	return &App{
		logger,
		storage,
	}
}

func (a *App) CreateEvent(ctx context.Context, id, title string) error {
	// TODO
	return nil
	// return a.storage.CreateEvent(storage.Event{ID: id, Title: title})
}

// TODO
