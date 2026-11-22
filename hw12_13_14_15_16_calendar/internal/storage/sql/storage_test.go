package sqlstorage

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage"
)

const (
	testUserID      = 1
	otherTestUserID = 2
)

func TestAdd(t *testing.T) {
	if _, exists := os.LookupEnv("TEST_CALENDAR_DSN"); !exists {
		t.Skip("no TEST_CALENDAR_DSN")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var err error

	s := New(os.Getenv("TEST_CALENDAR_DSN"))
	require.NoError(t, s.Connect(ctx))
	defer s.Close(ctx)

	event := storage.NewEvent()
	event.UserID = testUserID
	event.OnDate = time.Now()

	err = s.Add(ctx, *event)
	require.NoError(t, err)
	require.NotEmpty(t, event.ID, "event did not get id")

	foundEvent, err := s.GetByID(context.Background(), event.ID)
	require.NoError(t, err)

	event.OnDate = event.OnDate.UTC().Truncate(time.Microsecond)
	foundEvent.OnDate = foundEvent.OnDate.UTC().Truncate(time.Microsecond)

	require.EqualValues(t, event, foundEvent)

	event1 := storage.NewEvent()
	event1.UserID = testUserID
	require.ErrorIs(t, s.Add(context.Background(), *event1), storage.ErrOnDateRequired)

	require.NoError(t, s.DeleteAll(ctx))
}

func TestUpdate(t *testing.T) {
	if _, exists := os.LookupEnv("TEST_CALENDAR_DSN"); !exists {
		t.Skip("no TEST_CALENDAR_DSN")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var err error

	s := New(os.Getenv("TEST_CALENDAR_DSN"))
	require.NoError(t, s.Connect(ctx))
	defer s.Close(ctx)

	event := storage.NewEvent()
	event.UserID = testUserID
	event.Title = "update A"
	event.OnDate = time.Now()

	err = s.Add(context.Background(), *event)
	require.NoError(t, err)

	event.Title = "udpate B"

	err = s.Update(context.Background(), *event)
	require.NoError(t, err)

	foundEvent, err := s.GetByID(context.Background(), event.ID)

	event.OnDate = event.OnDate.UTC().Truncate(time.Microsecond)
	foundEvent.OnDate = foundEvent.OnDate.UTC().Truncate(time.Microsecond)

	require.NoError(t, err)
	require.EqualValues(t, event, foundEvent)

	require.NoError(t, s.DeleteAll(ctx))
}

func TestDelete(t *testing.T) {
	if _, exists := os.LookupEnv("TEST_CALENDAR_DSN"); !exists {
		t.Skip("no TEST_CALENDAR_DSN")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var err error

	s := New(os.Getenv("TEST_CALENDAR_DSN"))
	require.NoError(t, s.Connect(ctx))
	defer s.Close(ctx)

	event := storage.NewEvent()
	event.UserID = testUserID
	event.OnDate = time.Now()

	err = s.Add(ctx, *event)
	require.NoError(t, err)

	err = s.Delete(ctx, event.ID)
	require.NoError(t, err)

	_, err = s.GetByID(ctx, event.ID)
	require.ErrorIs(t, err, storage.ErrNotFound)

	require.NoError(t, s.DeleteAll(ctx))
}

func TestGetAll(t *testing.T) {
	if _, exists := os.LookupEnv("TEST_CALENDAR_DSN"); !exists {
		t.Skip("no TEST_CALENDAR_DSN")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var err error

	s := New(os.Getenv("TEST_CALENDAR_DSN"))
	require.NoError(t, s.Connect(ctx))
	defer s.Close(ctx)

	expectedEvents := make(map[string]string)

	for i := range 10 {
		event := storage.NewEvent()
		event.UserID = testUserID
		event.OnDate = time.Now().Add(time.Hour * time.Duration(i))
		event.Duration = time.Minute * 30

		expectedEvents[event.ID] = event.ID

		err := s.Add(context.Background(), *event)
		require.NoError(t, err)
	}

	events, err := s.GetAll(context.Background())
	require.NoError(t, err)

	require.GreaterOrEqual(t, len(events), 10)

	// проверяем что в хранилище разные события, а не одно и тоже или не nil
	ids := make(map[string]string)
	for _, e := range events {
		_, exists := expectedEvents[e.ID]
		if exists {
			ids[e.ID] = e.ID
		}
	}

	require.Equal(t, len(expectedEvents), len(ids))

	require.NoError(t, s.DeleteAll(ctx))
}

func TestErrors(t *testing.T) {
	if _, exists := os.LookupEnv("TEST_CALENDAR_DSN"); !exists {
		t.Skip("no TEST_CALENDAR_DSN")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	var err error

	s := New(os.Getenv("TEST_CALENDAR_DSN"))
	require.NoError(t, s.Connect(ctx))
	defer s.Close(ctx)

	// пересечение событий для пользователя
	t.Run("Add method generate ErrDateBusy for user", func(t *testing.T) {
		e := storage.NewEvent()
		e.OnDate = time.Now().Add(time.Hour * -12)
		e.Duration = time.Hour * 24
		e.UserID = 1

		require.NoError(t, s.Add(ctx, *e))

		eNext := storage.NewEvent()
		eNext.OnDate = time.Now().Add(time.Hour * -3)
		eNext.Duration = time.Minute * 30
		eNext.UserID = 1

		err = s.Add(ctx, *eNext)
		require.NotEqual(t, e.ID, eNext.ID, "ids must be different")
		require.ErrorIs(t, err, storage.ErrDateBusy)
		require.NoError(t, s.DeleteAll(ctx))
	})

	t.Run("Update method generate ErrNotFound by not existing id", func(t *testing.T) {
		e := storage.NewEvent()
		e.OnDate = time.Now().Add(time.Hour * -12)
		e.Duration = time.Hour * 24
		e.UserID = 1

		err = s.Update(ctx, *e)
		require.ErrorIs(t, err, storage.ErrNotFound)
	})

	t.Run("Delete method generate ErrNotFound by not existing id", func(t *testing.T) {
		err = s.Delete(ctx, uuid.Must(uuid.NewV7()).String())
		require.ErrorIs(t, err, storage.ErrNotFound)
	})

	t.Run("Update method generate ErrDateBusy for user", func(t *testing.T) {
		e := storage.NewEvent()
		e.OnDate = time.Now().Add(time.Hour * -12)
		e.Duration = time.Hour * 24
		e.UserID = 1

		require.NoError(t, s.Add(ctx, *e))

		eNext := storage.NewEvent()
		// через 3 дня
		eNext.OnDate = time.Now().Add(time.Hour * 3 * 24)
		eNext.Duration = time.Minute * 30
		eNext.UserID = 1

		require.NoError(t, s.Add(ctx, *eNext))

		eNext.OnDate = time.Now()
		err = s.Update(ctx, *eNext)
		require.ErrorIs(t, err, storage.ErrDateBusy)

		require.NoError(t, s.DeleteAll(ctx))
	})

	t.Run("no ErrDateBusy for different users", func(t *testing.T) {
		e := storage.NewEvent()
		e.OnDate = time.Now().Add(time.Hour * -12)
		e.Duration = time.Hour * 24
		e.UserID = testUserID

		require.NoError(t, s.Add(ctx, *e))

		eNext := storage.NewEvent()
		eNext.OnDate = time.Now().Add(time.Hour * -3)
		eNext.Duration = time.Minute * 30
		eNext.UserID = otherTestUserID

		require.NotEqual(t, e.ID, eNext.ID, "ids must be different")
		require.NoError(t, s.Add(ctx, *eNext))
		require.NoError(t, s.DeleteAll(ctx))
	})

	t.Run("no ErrDateBusy for close dates", func(t *testing.T) {
		e := storage.NewEvent()
		e.OnDate = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
		e.Duration = time.Hour * 24
		e.UserID = testUserID

		require.NoError(t, s.Add(context.Background(), *e))

		eNext := storage.NewEvent()
		eNext.OnDate = time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
		eNext.Duration = time.Minute * 30
		eNext.UserID = testUserID

		require.NotEqual(t, e.ID, eNext.ID, "ids must be different")
		require.NoError(t, s.Add(context.Background(), *eNext))

		require.NoError(t, s.DeleteAll(ctx))
	})
}

func TestRequirements(t *testing.T) {
	if _, exists := os.LookupEnv("TEST_CALENDAR_DSN"); !exists {
		t.Skip("no TEST_CALENDAR_DSN")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	s := New(os.Getenv("TEST_CALENDAR_DSN"))
	require.NoError(t, s.Connect(ctx))
	defer s.Close(ctx)

	t.Run("require UserId", func(t *testing.T) {
		event := storage.NewEvent()

		err := s.Add(context.Background(), *event)
		require.ErrorIs(t, err, storage.ErrUserIDRequired)
	})

	t.Run("require OnDate", func(t *testing.T) {
		event := storage.NewEvent()

		err := s.Add(context.Background(), *event)
		require.ErrorIs(t, err, storage.ErrUserIDRequired)
	})
}
