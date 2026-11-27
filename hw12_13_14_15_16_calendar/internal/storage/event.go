package storage

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNotFound         = errors.New("event not found")
	ErrUserIDRequired   = errors.New("userId required")
	ErrOnDateRequired   = errors.New("onDate required")
	ErrDateBusy         = errors.New("date is busy")
	ErrDurationRequired = errors.New("duration required")
)

type Event struct {
	ID               string        `db:"id" json:"id"`
	Title            string        `db:"title" json:"title"`
	OnDate           time.Time     `db:"on_date" json:"onDate"`
	Duration         time.Duration `db:"duration" json:"duration"`
	Description      string        `db:"description" json:"description"`
	UserID           int           `db:"user_id" json:"userId"`
	NotificateBefore time.Duration `db:"notificate_before" json:"notificateBefore"`
}

func NewEvent() *Event {
	// uuid.Must паникует при ошибке, не уверен стоит это использовать или нет
	return &Event{ID: uuid.Must(uuid.NewV7()).String()}
}
