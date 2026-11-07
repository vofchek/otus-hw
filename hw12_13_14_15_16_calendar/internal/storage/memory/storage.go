package memorystorage

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	storage "github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage"
)

// храним сериализованный мапу событий сериализованных в виде json
type Storage struct {
	storage map[string][]byte
	mu      sync.RWMutex //nolint:unused
}

func New() *Storage {
	return &Storage{
		mu:      sync.RWMutex{},
		storage: make(map[string][]byte),
	}
}

func (s *Storage) Add(ctx context.Context, event *storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("cant marshal event")
	}
	s.storage[event.ID] = d
	return nil
}

func (s *Storage) Update(ctx context.Context, event storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(event.ID) == 0 {
		return fmt.Errorf("event id is empty")
	}

	d, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("cant marshal event")
	}
	s.storage[event.ID] = d
	return nil
}

func (s *Storage) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.storage, id)
	return nil
}

func (s *Storage) GetById(ctx context.Context, id string) (*storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	d, ok := s.storage[id]
	if !ok {
		return nil, storage.ErrNotFound
	}

	var event storage.Event
	if err := json.Unmarshal(d, &event); err != nil {
		return nil, err
	}

	return &event, nil
}

func (s *Storage) GetAll(ctx context.Context) ([]*storage.Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	r := make([]*storage.Event, len(s.storage))
	i := 0
	for _, v := range s.storage {
		var event *storage.Event
		if err := json.Unmarshal(v, &event); err != nil {
			return nil, err
		}

		r[i] = event
	}
	return r, nil
}
