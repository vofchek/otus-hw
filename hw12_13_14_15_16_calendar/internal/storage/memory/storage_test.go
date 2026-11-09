package memorystorage

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	storage "github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage"
	"go.uber.org/goleak"
)

const testUserID = 1

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestStorage(t *testing.T) {
	t.Parallel()

	t.Run("add event", func(t *testing.T) {
		s := New()

		event := storage.NewEvent()
		event.UserID = testUserID
		event.OnDate = time.Now()

		err := s.Add(context.Background(), *event)
		require.NoError(t, err)

		require.NotEmpty(t, event.ID, "event did not get id")

		foundEvent, err := s.GetByID(context.Background(), event.ID)
		require.NoError(t, err)
		require.EqualValues(t, event, foundEvent)

		event1 := storage.NewEvent()
		event1.UserID = testUserID
		require.ErrorIs(t, s.Add(context.Background(), *event1), storage.ErrOnDateRequired)
	})

	t.Run("update event", func(t *testing.T) {
		s := New()

		event := storage.NewEvent()
		event.UserID = testUserID
		event.Title = "A"
		event.OnDate = time.Now()

		err := s.Add(context.Background(), *event)
		require.NoError(t, err)

		event.Title = "B"

		err = s.Update(context.Background(), *event)
		require.NoError(t, err)

		foundEvent, err := s.GetByID(context.Background(), event.ID)
		require.NoError(t, err)
		require.EqualValues(t, event, foundEvent)
	})

	t.Run("delete event", func(t *testing.T) {
		s := New()

		event := storage.NewEvent()
		event.UserID = testUserID
		event.OnDate = time.Now()
		err := s.Add(context.Background(), *event)
		require.NoError(t, err)

		err = s.Delete(context.Background(), event.ID)
		require.NoError(t, err)

		_, err = s.GetByID(context.Background(), event.ID)
		require.ErrorIs(t, err, storage.ErrNotFound)
	})

	t.Run("get all event", func(t *testing.T) {
		s := New()

		for i := range 10 {
			event := storage.NewEvent()
			event.UserID = testUserID
			event.OnDate = time.Now().Add(time.Hour * time.Duration(i))
			event.Duration = time.Minute * 30

			err := s.Add(context.Background(), *event)
			require.NoError(t, err)
		}

		events, err := s.GetAll(context.Background())
		require.NoError(t, err)

		require.Equal(t, 10, len(events))

		// проверяем что в хранилище разные события, а не одно и тоже или не nil
		ids := make(map[string]string)
		for _, e := range events {
			ids[e.ID] = e.ID
		}

		require.Equal(t, 10, len(ids))
	})

	// пересечение событий для пользователя
	t.Run("Add method generate ErrDateBusy for user", func(t *testing.T) {
		s := New()

		e := storage.NewEvent()
		e.OnDate = time.Now().Add(time.Hour * -12)
		e.Duration = time.Hour * 24
		e.UserID = 1

		require.NoError(t, s.Add(context.Background(), *e))

		eNext := storage.NewEvent()
		eNext.OnDate = time.Now().Add(time.Hour * -3)
		eNext.Duration = time.Minute * 30
		eNext.UserID = 1

		err := s.Add(context.Background(), *eNext)
		require.NotEqual(t, e.ID, eNext.ID, "ids must be different")
		require.ErrorIs(t, err, storage.ErrDateBusy)
	})

	t.Run("Update method generate ErrDateBusy for user", func(t *testing.T) {
		s := New()

		e := storage.NewEvent()
		e.OnDate = time.Now().Add(time.Hour * -12)
		e.Duration = time.Hour * 24
		e.UserID = 1

		require.NoError(t, s.Add(context.Background(), *e))

		eNext := storage.NewEvent()
		// через 3 дня
		eNext.OnDate = time.Now().Add(time.Hour * 3 * 24)
		eNext.Duration = time.Minute * 30
		eNext.UserID = 1

		require.NoError(t, s.Add(context.Background(), *eNext))

		eNext.OnDate = time.Now()
		err := s.Update(context.Background(), *eNext)
		require.ErrorIs(t, err, storage.ErrDateBusy)
	})

	t.Run("no ErrDateBusy for different users", func(t *testing.T) {
		s := New()

		e := storage.NewEvent()
		e.OnDate = time.Now().Add(time.Hour * -12)
		e.Duration = time.Hour * 24
		e.UserID = 1

		require.NoError(t, s.Add(context.Background(), *e))

		eNext := storage.NewEvent()
		eNext.OnDate = time.Now().Add(time.Hour * -3)
		eNext.Duration = time.Minute * 30
		eNext.UserID = 2

		require.NotEqual(t, e.ID, eNext.ID, "ids must be different")
		require.NoError(t, s.Add(context.Background(), *eNext))
	})
}

func TestRequirements(t *testing.T) {
	t.Parallel()

	t.Run("require UserId", func(t *testing.T) {
		s := New()

		event := storage.NewEvent()

		err := s.Add(context.Background(), *event)
		require.ErrorIs(t, err, storage.ErrUserIDRequired)
	})
}

func TestConcurrentOperations(t *testing.T) {
	t.Parallel()

	s := New()

	wg := sync.WaitGroup{}

	eventTime := time.Now()
	for i := range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			event := storage.NewEvent()
			event.UserID = testUserID + i
			event.OnDate = eventTime

			err := s.Add(context.Background(), *event)
			assert.NoError(t, err)

			event.Title = "this is new title"
			s.Update(context.Background(), *event)

			_, _ = s.GetAll(context.Background())
			_, _ = s.GetByID(context.Background(), event.ID)
		}()
	}
	wg.Wait()

	events, err := s.GetAll(context.Background())
	require.NoError(t, err, "error while getting all events")
	require.Len(t, events, 10, "must be 10 events")

	ids := make(map[string]struct{}, len(events))
	for _, e := range events {
		ids[e.ID] = struct{}{}
	}
	require.Len(t, ids, 10)
}
