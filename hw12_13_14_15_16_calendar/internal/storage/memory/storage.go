package memorystorage

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	storage "github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage"
)

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

func (s *Storage) Add(event *storage.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	event.ID = strconv.FormatInt(time.Now().UnixMilli(), 16)

	d, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("cant marshal event")
	}
	s.storage[event.ID] = d
	return nil
}

func (s *Storage) Update(event storage.Event) error {
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

func (s *Storage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.storage, id)
	return nil
}

func (s *Storage) GetById(id string) (*storage.Event, error) {
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

func (s *Storage) GetAll() ([]*storage.Event, error) {
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
