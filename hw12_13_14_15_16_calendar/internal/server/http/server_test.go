package internalhttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/app"
	"github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage"
	memorystorage "github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage/memory"
)

const (
	createPath = "/create"
	updatePath = "/update"
	deletePath = "/delete"
	dayPath    = "/get-on-date"
	weekPath   = "/get-on-week"
	monthPath  = "/get-on-month"

	testUserID = 1
)

type apiResponse[T any] struct {
	Success bool   `json:"success"`
	Data    T      `json:"data"`
	Error   string `json:"error"`
}

type testLogger struct{}

func (testLogger) Info(string)  {}
func (testLogger) Error(string) {}

type recordingLogger struct {
	messages []string
}

func (l *recordingLogger) Info(msg string) {
	l.messages = append(l.messages, msg)
}

func (l *recordingLogger) Error(string) {}

type testEnv struct {
	server *Server
	repo   *memorystorage.Storage
}

func newTestEnv(tb testing.TB, logg Logger) *testEnv {
	tb.Helper()

	if logg == nil {
		logg = testLogger{}
	}

	repo := memorystorage.New()
	return &testEnv{
		server: NewServer(logg, app.New(logg, repo)),
		repo:   repo,
	}
}

func doRequest(tb testing.TB, env *testEnv, method, path string, userID int, body any) *httptest.ResponseRecorder {
	tb.Helper()

	raw, err := json.Marshal(body)
	require.NoError(tb, err)

	req := httptest.NewRequest(method, path, bytes.NewReader(raw))
	req.Header.Set("User-Agent", "calendar-test")
	req.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))

	rec := httptest.NewRecorder()
	env.server.routes().ServeHTTP(rec, req)
	return rec
}

func decodeResponse[T any](tb testing.TB, rec *httptest.ResponseRecorder) apiResponse[T] {
	tb.Helper()

	var resp apiResponse[T]
	require.NoError(tb, json.Unmarshal(rec.Body.Bytes(), &resp))
	return resp
}

func createEvent(tb testing.TB, env *testEnv, title string, date time.Time, userID int) string {
	tb.Helper()

	rec := doRequest(tb, env, http.MethodPost, createPath, userID, CreateRequest{
		Title:    title,
		OnDate:   date,
		Duration: 60,
	})
	require.Equal(tb, http.StatusOK, rec.Code)

	resp := decodeResponse[storage.Event](tb, rec)
	require.True(tb, resp.Success)
	require.NotEmpty(tb, resp.Data.ID)
	return resp.Data.ID
}

func TestCreateHandler(t *testing.T) {
	env := newTestEnv(t, nil)
	date := time.Date(2030, time.March, 15, 10, 30, 0, 0, time.UTC)
	body := CreateRequest{
		Title:            "meeting",
		OnDate:           date,
		Duration:         3600,
		Description:      "discuss plans",
		NotificateBefore: time.Hour,
	}

	rec := doRequest(t, env, http.MethodPost, createPath, testUserID, body)
	require.Equal(t, http.StatusOK, rec.Code)

	resp := decodeResponse[storage.Event](t, rec)
	require.True(t, resp.Success)
	require.Empty(t, resp.Error)
	require.NotEmpty(t, resp.Data.ID)
	require.Equal(t, testUserID, resp.Data.UserID)
	require.Equal(t, body.Title, resp.Data.Title)
	require.Equal(t, body.Description, resp.Data.Description)
	require.Equal(t, body.NotificateBefore, resp.Data.NotificateBefore)
	require.Equal(t, time.Hour, resp.Data.Duration)
	require.True(t, date.Equal(resp.Data.OnDate))

	t.Run("busy date", func(t *testing.T) {
		rec := doRequest(t, env, http.MethodPost, createPath, testUserID, body)
		require.Equal(t, http.StatusInternalServerError, rec.Code)

		resp := decodeResponse[storage.Event](t, rec)
		require.False(t, resp.Success)
		require.Contains(t, resp.Error, storage.ErrDateBusy.Error())
	})

	t.Run("bad json", func(t *testing.T) {
		rec := doRequest(t, env, http.MethodPost, createPath, testUserID, "not an object")
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestUpdateHandler(t *testing.T) {
	env := newTestEnv(t, nil)
	oldDate := time.Date(2031, time.April, 1, 9, 0, 0, 0, time.UTC)
	eventID := createEvent(t, env, "old title", oldDate, 1000)

	newDate := oldDate.AddDate(0, 0, 1)
	body := UpdateRequest{
		ID: eventID,
		CreateRequest: CreateRequest{
			Title:            "new title",
			OnDate:           newDate,
			Duration:         120,
			Description:      "updated description",
			NotificateBefore: 2 * time.Hour,
		},
	}

	rec := doRequest(t, env, http.MethodPost, updatePath, testUserID, body)
	require.Equal(t, http.StatusOK, rec.Code)

	resp := decodeResponse[storage.Event](t, rec)
	require.True(t, resp.Success)
	require.Equal(t, eventID, resp.Data.ID)
	require.Equal(t, body.Title, resp.Data.Title)
	require.Equal(t, body.Description, resp.Data.Description)
	require.Equal(t, body.NotificateBefore, resp.Data.NotificateBefore)
	require.Equal(t, 2*time.Minute, resp.Data.Duration)
	require.True(t, newDate.Equal(resp.Data.OnDate))

	stored, err := env.repo.GetByID(t.Context(), eventID)
	require.NoError(t, err)
	require.Equal(t, body.Title, stored.Title)
	require.Equal(t, body.Description, stored.Description)
	require.Equal(t, body.NotificateBefore, stored.NotificateBefore)
	require.Equal(t, 2*time.Minute, stored.Duration)
}

func TestDeleteHandler(t *testing.T) {
	env := newTestEnv(t, nil)
	date := time.Date(2033, time.June, 1, 12, 0, 0, 0, time.UTC)
	eventID := createEvent(t, env, "to delete", date, 2000)

	rec := doRequest(t, env, http.MethodPost, deletePath, testUserID, DeleteRequest{ID: eventID})
	require.Equal(t, http.StatusOK, rec.Code)

	resp := decodeResponse[storage.Event](t, rec)
	require.True(t, resp.Success)

	_, err := env.repo.GetByID(t.Context(), eventID)
	require.ErrorIs(t, err, storage.ErrNotFound)

	t.Run("unknown id", func(t *testing.T) {
		rec := doRequest(t, env, http.MethodPost, deletePath, testUserID, DeleteRequest{ID: "missing"})
		require.Equal(t, http.StatusInternalServerError, rec.Code)

		resp := decodeResponse[storage.Event](t, rec)
		require.False(t, resp.Success)
		require.Contains(t, resp.Error, storage.ErrNotFound.Error())
	})
}

func TestGetEventsHandler(t *testing.T) {
	date := time.Date(2032, time.May, 12, 14, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		path string
	}{
		{name: "day", path: dayPath},
		{name: "week", path: weekPath},
		{name: "month", path: monthPath},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := newTestEnv(t, nil)
			eventID := createEvent(t, env, tc.name, date, 3000)

			other := storage.NewEvent()
			other.UserID = 100
			other.OnDate = date
			other.Duration = time.Minute
			require.NoError(t, env.repo.Add(t.Context(), *other))

			rec := doRequest(t, env, http.MethodGet, tc.path, 3000, GetRequest{Date: date})
			require.Equal(t, http.StatusOK, rec.Code)

			resp := decodeResponse[[]storage.Event](t, rec)
			require.True(t, resp.Success)
			require.Len(t, resp.Data, 1)
			require.Equal(t, eventID, resp.Data[0].ID)
		})
	}
}

func TestIndexHandler(t *testing.T) {
	env := newTestEnv(t, nil)

	rec := httptest.NewRecorder()
	env.server.routes().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	require.Equal(t, http.StatusOK, rec.Code)
	require.Contains(t, rec.Body.String(), "success")
}

func TestLoggingMiddleware(t *testing.T) {
	logg := &recordingLogger{}
	env := newTestEnv(t, logg)

	rec := doRequest(t, env, http.MethodGet, "/", testUserID, nil)
	require.Equal(t, http.StatusOK, rec.Code)
	require.Len(t, logg.messages, 1)

	msg := logg.messages[0]
	require.Contains(t, msg, http.MethodGet)
	require.Contains(t, msg, "200")
	require.Contains(t, msg, "HTTP/1.1")
	require.Contains(t, msg, "calendar-test")
}
