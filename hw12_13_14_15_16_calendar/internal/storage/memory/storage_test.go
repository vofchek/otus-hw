package memorystorage

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	storage "github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage"
)

func TestStorage(t *testing.T) {
	t.Run("add event", func(t *testing.T) {
		s := New()

		event := storage.NewEvent()

		err := s.Add(context.Background(), event)
		require.NoError(t, err)

		assert.NotEmpty(t, event.ID, "event did not get id")

		foundEvent, err := s.GetById(context.Background(), event.ID)
		assert.NoError(t, err)
		assert.EqualValues(t, event, foundEvent)
	})

	t.Run("update event", func(t *testing.T) {
		s := New()

		event := storage.NewEvent()
		event.Title = "A"

		err := s.Add(context.Background(), event)
		require.NoError(t, err)

		event.Title = "B"

		err = s.Update(context.Background(), *event)
		require.NoError(t, err)

		require.Equal(t, "B", event.Title)
	})

	t.Run("delete event", func(t *testing.T) {
		s := New()

		event := storage.NewEvent()
		err := s.Add(context.Background(), event)
		require.NoError(t, err)

		err = s.Delete(context.Background(), event.ID)
		require.NoError(t, err)

		_, err = s.GetById(context.Background(), event.ID)
		require.ErrorIs(t, storage.ErrNotFound, err)
	})

	t.Run("get all event", func(t *testing.T) {
		s := New()

		event := storage.NewEvent()
		err := s.Add(context.Background(), event)
		require.NoError(t, err)

		events, err := s.GetAll(context.Background())
		require.NoError(t, err)

		assert.Equal(t, 1, len(events))
	})
}
