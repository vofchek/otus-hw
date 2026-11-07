package sqlstorage

import (
	"context"
	"database/sql"

	"github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage"
)

type Storage struct {
	db   *sql.DB
	conn *sql.Conn
}

func New(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func (s *Storage) Connect(ctx context.Context) error {
	conn, err := s.db.Conn(ctx)
	s.conn = conn
	return err
}

func (s *Storage) Close(ctx context.Context) error {
	return s.db.Close()
}

func (s *Storage) Add(ctx context.Context, event *storage.Event) error {
	_, err := s.conn.ExecContext(
		ctx,
		`
			INSERT INTO calendar_event(id, title, on_date, duration, description, user_id, notificate_before)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`,
		event.ID,
		event.Title,
		event.OnDate,
		event.Duration,
		event.Description,
		event.UserId,
		event.NotificateBefore,
	)
	return err
}

func (s *Storage) Update(ctx context.Context, event storage.Event) error {
	_, err := s.conn.ExecContext(
		ctx,
		`
			UPDATE calendar_event
			SET
				title = $2,
				on_date = $3,
				duration = $4,
				description = $5,
				user_id = $6,
				notificate_before = $7
			WHERE id = $1
		`,
		event.ID,
		event.Title,
		event.OnDate,
		event.Duration,
		event.Description,
		event.UserId,
		event.NotificateBefore,
	)
	return err
}

func (s *Storage) Delete(ctx context.Context, id string) error {
	_, err := s.conn.ExecContext(ctx, "DELETE FROM calendar_event WHERE id = $1", id)
	return err
}

func (s *Storage) GetById(ctx context.Context, id string) (*storage.Event, error) {
	var event *storage.Event
	err := s.conn.QueryRowContext(ctx, "SELECT * FROM calendar_event WHERE id = $1", id).Scan(event)
	return event, err
}

func (s *Storage) GetAll(ctx context.Context) ([]*storage.Event, error) {
	events := make([]*storage.Event, 0)
	rows, err := s.conn.QueryContext(ctx, "SELECT * FROM calendar_event ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var event storage.Event
		err := rows.Scan(&event)
		if err != nil {
			return nil, err
		}

		events = append(events, &event)
	}

	if rows.Err() != nil {
		return events, nil
	}

	return events, nil
}
