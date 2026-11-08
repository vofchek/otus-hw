package storage

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound       = errors.New("event not found")
	ErrUserIdRequired = errors.New("userId required")
	ErrOnDateRequired = errors.New("onDate required")
	ErrDateBusy       = errors.New("date is busy")
)

type Event struct {
	ID               string
	Title            string
	OnDate           time.Time
	Duration         time.Duration
	Description      string
	UserId           int
	NotificateBefore time.Duration
}

func NewEvent() *Event {
	// uuid.Must паникует при ошибке, не уверен стоит это использовать или нет
	return &Event{ID: uuid.Must(uuid.NewV7()).String()}
}
