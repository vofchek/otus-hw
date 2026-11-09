package storage

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound       = errors.New("event not found")
	ErrUserIDRequired = errors.New("userId required")
	ErrOnDateRequired = errors.New("onDate required")
	ErrDateBusy       = errors.New("date is busy")
)

type Event struct {
	ID               string        `db:"id"`
	Title            string        `db:"title"`
	OnDate           time.Time     `db:"on_date"`
	Duration         time.Duration `db:"duration"`
	Description      string        `db:"description"`
	UserID           int           `db:"user_id"`
	NotificateBefore time.Duration `db:"notificate_before"`
}

func NewEvent() *Event {
	// uuid.Must паникует при ошибке, не уверен стоит это использовать или нет
	return &Event{ID: uuid.Must(uuid.NewV7()).String()}
}
