package internalgrpc

import (
	"context"
	"fmt"
	"net"

	"github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/pb"
	"google.golang.org/grpc"
)

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Server struct {
	Logger Logger
	App    Application
	server *grpc.Server
}

func NewServer(logger Logger, app Application) *Server {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(LogInterceptor(logger)),
	)
	pb.RegisterEventsServiceServer(server, NewService(app))

	return &Server{
		Logger: logger,
		App:    app,
		server: server,
	}
}

// address - хост + порт, чтобы не передавть эти данные через контекст.
func (s *Server) Start(_ context.Context, address string) error {
	s.Logger.Info(fmt.Sprintf("Listen on: %s", address))
	lsn, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	return s.server.Serve(lsn)
}

func (s *Server) Stop(_ context.Context) error {
	s.Logger.Info("Stopping server")
	s.server.GracefulStop()
	return nil
}
