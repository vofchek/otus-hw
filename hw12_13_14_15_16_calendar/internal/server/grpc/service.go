package internalgrpc

import (
	"context"
	"time"

	"github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/pb"
	"github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const userID = 1

type Service struct {
	App Application

	pb.UnimplementedEventsServiceServer
}

type Application interface {
	CreateEvent(
		ctx context.Context,
		userID int,
		title string,
		onDate time.Time,
		duration time.Duration,
		description string,
		notificateBefore time.Duration,
	) (storage.Event, error)

	UpdateEvent(
		ctx context.Context,
		id string,
		title string,
		onDate time.Time,
		duration time.Duration,
		description string,
		notificateBefore time.Duration,
	) (storage.Event, error)

	DeleteEvent(ctx context.Context, id string) error

	GetOnDate(ctx context.Context, userID int, date time.Time) ([]storage.Event, error)
	GetOnWeek(ctx context.Context, userID int, date time.Time) ([]storage.Event, error)
	GetOnMonth(ctx context.Context, userID int, date time.Time) ([]storage.Event, error)
}

func NewService(app Application) *Service {
	return &Service{
		App: app,
	}
}

func mapEventToEventMessage(e *storage.Event) *pb.Event {
	return &pb.Event{
		Id:               e.ID,
		Title:            e.Title,
		OnDate:           timestamppb.New(e.OnDate),
		Duration:         int64(e.Duration),
		Description:      e.Description,
		UserId:           int64(e.UserID),
		NotificateBefore: int64(e.NotificateBefore),
	}
}

func (s *Service) Create(ctx context.Context, req *pb.CreateRequest) (*pb.CreateResponse, error) {
	e, err := s.App.CreateEvent(
		ctx,
		userID,
		req.Title,
		req.OnDate.AsTime(),
		time.Duration(req.Duration),
		req.Description,
		time.Duration(req.NotificateBefore),
	)
	if err != nil {
		return nil, err
	}

	resp := &pb.CreateResponse{
		Data: mapEventToEventMessage(&e),
	}

	return resp, nil
}

func (s *Service) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	e, err := s.App.UpdateEvent(
		ctx,
		req.Id,
		req.Title,
		req.OnDate.AsTime(),
		time.Duration(req.Duration),
		req.Description,
		time.Duration(req.NotificateBefore),
	)
	if err != nil {
		return nil, err
	}

	resp := &pb.UpdateResponse{
		Data: mapEventToEventMessage(&e),
	}

	return resp, nil
}

func (s *Service) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	err := s.App.DeleteEvent(ctx, req.Id)
	if err != nil {
		return nil, err
	}

	resp := &pb.DeleteResponse{}
	return resp, nil
}

func (s *Service) GetOnDate(ctx context.Context, req *pb.GetOnDateRequest) (*pb.GetOnDateResponse, error) {
	resp := &pb.GetOnDateResponse{}
	events, err := s.App.GetOnDate(ctx, userID, req.Date.AsTime())
	if err != nil {
		return nil, err
	}

	resp.Data = make([]*pb.Event, len(events))
	for i, e := range events {
		resp.Data[i] = mapEventToEventMessage(&e)
	}

	return resp, nil
}

func (s *Service) GetOnWeek(ctx context.Context, req *pb.GetOnWeekRequest) (*pb.GetOnWeekResponse, error) {
	resp := &pb.GetOnWeekResponse{}
	events, err := s.App.GetOnWeek(ctx, userID, req.Date.AsTime())
	if err != nil {
		return nil, err
	}

	resp.Data = make([]*pb.Event, len(events))
	for i, e := range events {
		resp.Data[i] = mapEventToEventMessage(&e)
	}

	return resp, nil
}

func (s *Service) GetOnMonth(ctx context.Context, req *pb.GetOnMonthRequest) (*pb.GetOnMonthResponse, error) {
	resp := &pb.GetOnMonthResponse{}
	events, err := s.App.GetOnMonth(ctx, userID, req.Date.AsTime())
	if err != nil {
		return nil, err
	}

	resp.Data = make([]*pb.Event, len(events))
	for i, e := range events {
		resp.Data[i] = mapEventToEventMessage(&e)
	}

	return resp, nil
}
