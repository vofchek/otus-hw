package internalhttp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

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
		server: &http.Server{
			ReadHeaderTimeout: time.Second,
		},
	}
}

// address - хост + порт, чтобы не передавть эти данные через контекст.
func (s *Server) Start(ctx context.Context, address string) error {
	mux := http.NewServeMux()
	mux.Handle("GET /", loggingMiddleware(s.Logger, http.HandlerFunc(s.indexHandler)))

	s.server.Addr = address
	s.server.Handler = mux

	// для отлова ошибок прослушки сокета.
	// используется, чтобы убрать os.Exit.
	errCh := make(chan error)
	defer close(errCh)

	go func() {
		s.Logger.Info(fmt.Sprintf("Listen on: %s", address))
		err := s.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("cant start server, err: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errCh:
		return err
	}
}

func (s *Server) Stop(ctx context.Context) error {
	s.Logger.Info("Stopping server")
	err := s.server.Shutdown(ctx)
	return err
}

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	_ = r.Context()
	t, err := json.Marshal(`{success: true, payload: [1,2,3]}`)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("json decode failed"))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(t)
}
