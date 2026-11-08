package internalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
)

var (
	ErrConfigAddressEmpty       = errors.New("address is empty")
	ErrConfigAddressIsNotString = errors.New("address is not string")
)

const ContextAddressKey = "http.address"

type Server struct {
	Logger Logger
	App    Application
	server *http.Server
}

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type Application interface { // TODO
}

func NewServer(logger Logger, app Application) *Server {
	return &Server{
		Logger: logger,
		App:    app,
	}
}

func (s *Server) Start(ctx context.Context) error {
	address := ctx.Value(ContextAddressKey)
	if address == nil {
		return ErrConfigAddressEmpty
	}
	address, ok := address.(string)
	if !ok {
		return ErrConfigAddressIsNotString
	}

	mux := http.NewServeMux()
	mux.Handle("GET /", loggingMiddleware(s.Logger, http.HandlerFunc(indexHandler)))

	s.server = &http.Server{
		Addr:    address.(string),
		Handler: mux,
	}

	go func() {
		err := s.server.ListenAndServe()
		if err != nil {
			s.Logger.Error(fmt.Sprintf("cant start server, err: %v", err))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	err := s.server.Shutdown(ctx)
	return err
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	t, err := json.Marshal(`{success: true, payload: [1,2,3]}\n`)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("json decode failed"))
		return
	}

	w.Write(t)
}
