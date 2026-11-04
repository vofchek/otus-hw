package memorystorage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	storage "github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage"
)

func TestStorage(t *testing.T) {
	t.Run("add event", func(t *testing.T) {
		s := New()

		event := storage.Event{}

		err := s.Add(&event)
		require.NoError(t, err)

		assert.NotEmpty(t, event.ID, "event did not get id")

		foundEvent, err := s.GetById(event.ID)
		assert.NoError(t, err)
		assert.EqualValues(t, event, *foundEvent)
	})

	t.Run("update event", func(t *testing.T) {
		s := New()

		event := storage.Event{
			Title: "A",
		}

		err := s.Add(&event)
		require.NoError(t, err)

		event.Title = "B"

		err = s.Update(event)
		require.NoError(t, err)

		require.Equal(t, "B", event.Title)
	})

	t.Run("delete event", func(t *testing.T) {
		s := New()

		event := &storage.Event{}
		err := s.Add(event)
		require.NoError(t, err)

		err = s.Delete(event.ID)
		require.NoError(t, err)

		_, err = s.GetById(event.ID)
		require.ErrorIs(t, storage.ErrNotFound, err)
	})

	t.Run("get all event", func(t *testing.T) {
		s := New()

		event := &storage.Event{}
		err := s.Add(event)
		require.NoError(t, err)

		events, err := s.GetAll()
		require.NoError(t, err)

		assert.Equal(t, 1, len(events))
	})
}
