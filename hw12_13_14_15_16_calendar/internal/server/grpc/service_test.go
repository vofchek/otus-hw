package internalgrpc

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/app"
	"github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/logger"
	"github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/pb"
	"github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage/memory"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const testUserID = 1

func userContext(tb testing.TB) context.Context {
	tb.Helper()
	return metadata.NewIncomingContext(
		tb.Context(),
		metadata.Pairs("user-id", strconv.Itoa(testUserID)),
	)
}

func createService(tb testing.TB, storage *memorystorage.Storage) *Service {
	tb.Helper()

	logg, err := logger.New("info", "")
	require.NoError(tb, err)

	if storage == nil {
		storage = memorystorage.New()
	}

	return NewService(app.New(logg, storage))
}

func provideEvent(tb testing.TB, service *Service, title string, date time.Time, duration int64) (id string) {
	tb.Helper()

	req := &pb.CreateRequest{
		Title:    title,
		OnDate:   timestamppb.New(date),
		Duration: duration,
	}
	resp, err := service.Create(userContext(tb), req)
	require.NoError(tb, err, "error on create")
	require.NotEmpty(tb, resp.Data.Id)

	return resp.Data.Id
}

func TestCreate(t *testing.T) {
	s := createService(t, nil)

	t.Run("normal data", func(t *testing.T) {
		req := &pb.CreateRequest{
			Title:    "test",
			OnDate:   timestamppb.New(time.Now()),
			Duration: 100,
		}
		resp, err := s.Create(userContext(t), req)
		require.NoError(t, err, "error on create")

		require.NotEmpty(t, resp.Data.Id)
	})
}

func TestUpdate(t *testing.T) {
	s := createService(t, nil)

	eventID := provideEvent(t, s, "1", time.Now(), 100)

	t.Run("normal data", func(t *testing.T) {
		newDate := time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC)

		req := &pb.UpdateRequest{
			Id:       eventID,
			Title:    "test updated",
			OnDate:   timestamppb.New(newDate),
			Duration: int64(500),
		}
		resp, err := s.Update(userContext(t), req)
		require.NoError(t, err, "error on update")

		require.Equal(t, "test updated", resp.Data.Title)
		require.Equal(
			t,
			newDate.Truncate(time.Second),
			resp.Data.OnDate.AsTime().Truncate(time.Second),
		)
		require.Equal(t, int64(500), resp.Data.Duration)
	})
}

func TestDelete(t *testing.T) {
	repo := memorystorage.New()
	service := createService(t, repo)

	eventID := provideEvent(t, service, "1", time.Now(), 10000)

	t.Run("delete existing event", func(t *testing.T) {
		req := &pb.DeleteRequest{
			Id: eventID,
		}
		_, err := service.Delete(userContext(t), req)
		require.NoError(t, err, "error on delete")

		_, err = repo.GetByID(userContext(t), eventID)
		require.ErrorIs(t, err, storage.ErrNotFound)
	})
}

func TestGetEvents(t *testing.T) {
	service := createService(t, nil)

	t.Run("get on date", func(t *testing.T) {
		date := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		eventID := provideEvent(t, service, "date", date, 5000)

		req := &pb.GetOnDateRequest{
			Date: timestamppb.New(date),
		}

		events, err := service.GetOnDate(userContext(t), req)
		require.NoError(t, err)
		for _, e := range events.Data {
			require.Equal(t, eventID, e.Id)
		}
	})

	t.Run("get on week", func(t *testing.T) {
		date := time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
		eventID := provideEvent(t, service, "week", date, 6000)

		req := &pb.GetOnWeekRequest{
			Date: timestamppb.New(date),
		}

		events, err := service.GetOnWeek(userContext(t), req)
		require.NoError(t, err)
		for _, e := range events.Data {
			require.Equal(t, eventID, e.Id)
		}
	})

	t.Run("get on month", func(t *testing.T) {
		date := time.Date(2002, 1, 1, 0, 0, 0, 0, time.UTC)
		eventID := provideEvent(t, service, "month", date, 3000)

		req := &pb.GetOnMonthRequest{
			Date: timestamppb.New(date),
		}

		events, err := service.GetOnMonth(userContext(t), req)
		require.NoError(t, err)
		for _, e := range events.Data {
			require.Equal(t, eventID, e.Id)
		}
	})
}
