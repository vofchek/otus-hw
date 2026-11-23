package memorystorage

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	storage "github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage"
)

// храним мапу объектов событий.
type Storage struct {
	storage map[string]storage.Event
	mu      sync.RWMutex
}

func New() *Storage {
	return &Storage{
		mu:      sync.RWMutex{},
		storage: make(map[string]storage.Event),
	}
}

func (s *Storage) Connect(_ context.Context) error {
	return nil
}

func (s *Storage) Close(_ context.Context) error {
	return nil
}

func (s *Storage) Add(_ context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.assertEventValid(&event)
	if err != nil {
		return err
	}

	s.storage[event.ID] = event
	return nil
}

func (s *Storage) Update(_ context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.assertEventValid(&event)
	if err != nil {
		return err
	}

	_, exists := s.storage[event.ID]
	if !exists {
		return storage.ErrNotFound
	}

	s.storage[event.ID] = event
	return nil
}

func (s *Storage) Delete(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.storage[id]
	if !exists {
		return storage.ErrNotFound
	}

	delete(s.storage, id)
	return nil
}

func (s *Storage) GetByID(_ context.Context, id string) (*storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	event, ok := s.storage[id]
	if !ok {
		return nil, storage.ErrNotFound
	}

	return &event, nil
}

func (s *Storage) GetAll(_ context.Context) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	r := make([]storage.Event, len(s.storage))
	i := 0
	for _, v := range s.storage {
		r[i] = v
		i++
	}
	return r, nil
}

func (s *Storage) GetByPeriod(_ context.Context, userId int, from, to time.Time) ([]storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	r := make([]storage.Event, 0)
	for _, v := range s.storage {
		if v.OnDate.After(from) && v.OnDate.Before(to) && v.UserID == userId {
			r = append(r, v)
		}
	}
	slices.SortFunc(r, func(a, b storage.Event) int {
		return a.OnDate.Compare(b.OnDate)
	})
	return r, nil
}

func (s *Storage) assertEventValid(event *storage.Event) error {
	if len(event.ID) == 0 {
		return fmt.Errorf("event id is empty")
	}

	if event.UserID <= 0 {
		return storage.ErrUserIDRequired
	}

	if event.OnDate.IsZero() {
		return storage.ErrOnDateRequired
	}

	if event.Duration <= 0 {
		return storage.ErrDurationRequired
	}

	for _, e := range s.storage {
		if e.UserID != event.UserID || e.ID == event.ID {
			continue
		}

		eStart := e.OnDate
		eEnd := e.OnDate.Add(e.Duration)
		eventStart := event.OnDate
		eventEnd := event.OnDate.Add(event.Duration)

		intersected := eStart.Compare(eventEnd) < 0 && eEnd.Compare(eventStart) > 0
		if intersected {
			return fmt.Errorf(
				"%w, intersection with event: {ID: %s, OnDate: %s, Duration: %s}",
				storage.ErrDateBusy,
				e.ID,
				e.OnDate,
				e.Duration,
			)
		}
	}

	return nil
}
