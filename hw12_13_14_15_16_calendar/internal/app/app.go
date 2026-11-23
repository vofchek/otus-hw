package app

import (
	"context"
	"time"

	"github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage"
)

type App struct {
	logger  Logger
	storage Storage
}

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Storage interface {
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
	Add(ctx context.Context, event storage.Event) error
	Update(ctx context.Context, event storage.Event) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*storage.Event, error)
	GetAll(ctx context.Context) ([]storage.Event, error)
	GetByPeriod(ctx context.Context, userID int, from, to time.Time) ([]storage.Event, error)
}

func New(logger Logger, storage Storage) *App {
	return &App{
		logger,
		storage,
	}
}

func (a *App) CreateEvent(
	ctx context.Context, userID int, title string, onDate time.Time, duration time.Duration,
) (storage.Event, error) {
	event := storage.NewEvent()
	event.UserID = userID
	event.Title = title
	event.OnDate = onDate
	event.Duration = duration

	return *event, a.storage.Add(ctx, *event)
}

func (a *App) UpdateEvent(
	ctx context.Context, id string, title string, onDate time.Time, duration time.Duration,
) (storage.Event, error) {
	event, err := a.storage.GetByID(ctx, id)
	if err != nil {
		return storage.Event{}, err
	}

	event.Title = title
	event.OnDate = onDate
	event.Duration = duration

	return *event, a.storage.Update(ctx, *event)
}

func (a *App) DeleteEvent(ctx context.Context, id string) error {
	return a.storage.Delete(ctx, id)
}

func (a *App) GetOnDate(ctx context.Context, userID int, date time.Time) ([]storage.Event, error) {
	return a.storage.GetByPeriod(ctx, userID, date, date)
}

func (a *App) GetOnWeek(ctx context.Context, userID int, date time.Time) ([]storage.Event, error) {
	offset := (int(date.Weekday()) + 6) % 7
	from := date.AddDate(0, 0, -offset)
	to := date.AddDate(0, 0, 7-offset)
	return a.storage.GetByPeriod(ctx, userID, from, to)
}

func (a *App) GetOnMonth(ctx context.Context, userID int, date time.Time) ([]storage.Event, error) {
	from := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 1, -1)
	return a.storage.GetByPeriod(ctx, userID, from, to)
}
