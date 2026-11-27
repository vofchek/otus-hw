package internalhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage"
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
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("GET /", loggingMiddleware(s.Logger, http.HandlerFunc(s.indexHandler)))
	mux.Handle("POST /create", loggingMiddleware(s.Logger, http.HandlerFunc(s.createHandler)))
	mux.Handle("POST /update", loggingMiddleware(s.Logger, http.HandlerFunc(s.updateHandler)))
	mux.Handle("POST /delete", loggingMiddleware(s.Logger, http.HandlerFunc(s.deleteHandler)))
	mux.Handle("GET /get-on-date", loggingMiddleware(s.Logger, http.HandlerFunc(s.getOnDateHandler)))
	mux.Handle("GET /get-on-week", loggingMiddleware(s.Logger, http.HandlerFunc(s.getOnWeekHandler)))
	mux.Handle("GET /get-on-month", loggingMiddleware(s.Logger, http.HandlerFunc(s.getOnMonthHandler)))

	return mux
}

func (s *Server) Start(ctx context.Context, address string) error {
	s.server.Addr = address
	s.server.Handler = s.routes()

	// для отлова ошибок прослушки сокета.
	// используется, чтобы убрать os.Exit.
	errCh := make(chan error)

	go func() {
		s.Logger.Info(fmt.Sprintf("Listen on: %s", address))
		err := s.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("cant start server, err: %w", err)
		}
		close(errCh)
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

// index обработчик.
func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {
	_ = r.Context()

	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte(`{success: true}`))
}

// структура ответа.
type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data"`
	Error   string `json:"error"`
}

// формирует ответ и записывает http код.
func writeResponse(w http.ResponseWriter, data any, err error) {
	var resp Response
	if err != nil {
		resp = Response{Success: false, Error: err.Error()}
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		resp = Response{Success: true, Data: data}
	}

	respStr, jsonErr := json.Marshal(resp)
	if jsonErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(respStr)
}

func getUserID(r *http.Request) (int, error) {
	rawUserID := r.Header.Get("X-User-ID")
	if rawUserID == "" {
		return 0, errors.New("no userID in X-User-ID header")
	}

	return strconv.Atoi(rawUserID)
}

// обработчик создания.
type CreateRequest struct {
	Title            string        `json:"title"`
	OnDate           time.Time     `json:"onDate"`
	Duration         int64         `json:"duration"`
	Description      string        `json:"description"`
	NotificateBefore time.Duration `json:"notificateBefore"`
}

func (s *Server) createHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	userID, err := getUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unknown user"))
		return
	}

	buf := make([]byte, r.ContentLength)
	_, err = r.Body.Read(buf)
	if err != nil && errors.Is(err, io.EOF) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("cant read body request"))
		return
	}

	params := &CreateRequest{}
	err = json.Unmarshal(buf, params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(fmt.Appendf(nil, "bad json, json: %v", buf))
		return
	}

	e, err := s.App.CreateEvent(
		r.Context(),
		userID,
		params.Title,
		params.OnDate,
		time.Duration(params.Duration*int64(time.Second)),
		params.Description,
		params.NotificateBefore,
	)
	writeResponse(w, e, err)
}

// обработчик обновления.
type UpdateRequest struct {
	ID string `json:"id"`
	CreateRequest
}

func (s *Server) updateHandler(w http.ResponseWriter, r *http.Request) {
	buf := make([]byte, r.ContentLength)
	_, err := r.Body.Read(buf)
	if err != nil && errors.Is(err, io.EOF) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	params := &UpdateRequest{}
	err = json.Unmarshal(buf, params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	e, err := s.App.UpdateEvent(r.Context(),
		params.ID,
		params.Title,
		params.OnDate,
		time.Duration(params.Duration*int64(time.Second)),
		params.Description,
		params.NotificateBefore,
	)
	writeResponse(w, e, err)
}

// обработчик удаления.
type DeleteRequest struct {
	ID string `json:"id"`
}

func (s *Server) deleteHandler(w http.ResponseWriter, r *http.Request) {
	buf := make([]byte, r.ContentLength)
	_, err := r.Body.Read(buf)
	if err != nil && errors.Is(err, io.EOF) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	params := &UpdateRequest{}
	err = json.Unmarshal(buf, params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = s.App.DeleteEvent(r.Context(), params.ID)
	writeResponse(w, nil, err)
}

// обработчики получения списка по датам.
type GetRequest struct {
	Date time.Time `json:"date"`
}

func (s *Server) getOnDateHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	userID, err := getUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unknown user"))
		return
	}

	buf := make([]byte, r.ContentLength)
	_, err = r.Body.Read(buf)
	if err != nil && errors.Is(err, io.EOF) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	params := &GetRequest{}
	err = json.Unmarshal(buf, params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	events, err := s.App.GetOnDate(r.Context(), userID, params.Date)
	writeResponse(w, events, err)
}

func (s *Server) getOnWeekHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	userID, err := getUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unknown user"))
		return
	}

	buf := make([]byte, r.ContentLength)
	_, err = r.Body.Read(buf)
	if err != nil && errors.Is(err, io.EOF) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	params := &GetRequest{}
	err = json.Unmarshal(buf, params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	events, err := s.App.GetOnWeek(r.Context(), userID, params.Date)
	writeResponse(w, events, err)
}

func (s *Server) getOnMonthHandler(w http.ResponseWriter, r *http.Request) {
	var err error

	userID, err := getUserID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unknown user"))
		return
	}

	buf := make([]byte, r.ContentLength)
	_, err = r.Body.Read(buf)
	if err != nil && errors.Is(err, io.EOF) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	params := &GetRequest{}
	err = json.Unmarshal(buf, params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	events, err := s.App.GetOnMonth(r.Context(), userID, params.Date)
	writeResponse(w, events, err)
}
