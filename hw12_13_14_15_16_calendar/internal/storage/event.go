package storage

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("event not found")
)

type Event struct {
	ID               string
	Title            string
	OnDate           time.Time
	Duration         time.Duration
	Desc             string
	UserId           int
	NotificateBefore time.Duration
}
