package sqlstorage

import (
	"context"

	"github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage"
)

type Storage struct { // TODO
}

func New() *Storage {
	return &Storage{}
}

func (s *Storage) Connect(ctx context.Context) error {
	// TODO
	return nil
}

func (s *Storage) Close(ctx context.Context) error {
	// TODO
	return nil
}

func (s *Storage) Add(event *storage.Event) error {
	return nil
}

func (s *Storage) Update(event storage.Event) error {
	return nil
}

func (s *Storage) Delete(id string) error {
	return nil
}

func (s *Storage) GetById(id string) (*storage.Event, error) {
	return nil, nil
}

func (s *Storage) GetAll() ([]*storage.Event, error) {
	return make([]*storage.Event, 0), nil
}
