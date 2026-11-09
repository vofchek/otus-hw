package sqlstorage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/vofchek/otus-hw/hw12_13_14_15_calendar/internal/storage"
)

type Storage struct {
	dsn  string
	conn *pgx.Conn
}

func New(dsn string) *Storage {
	return &Storage{dsn: dsn}
}

func (s *Storage) Connect(ctx context.Context) error {
	conn, err := pgx.Connect(ctx, s.dsn)
	if err != nil {
		return err
	}

	if err := conn.Ping(ctx); err != nil {
		return err
	}

	s.conn = conn
	return nil
}

func (s *Storage) Close(ctx context.Context) error {
	return s.conn.Close(ctx)
}

func (s *Storage) Add(ctx context.Context, event storage.Event) error {
	if err := s.assertEventValid(ctx, event); err != nil {
		return err
	}

	_, err := s.conn.Exec(
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
		event.UserID,
		event.NotificateBefore,
	)
	return err
}

func (s *Storage) Update(ctx context.Context, event storage.Event) error {
	if err := s.assertEventValid(ctx, event); err != nil {
		return err
	}

	_, err := s.conn.Exec(
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
		event.UserID,
		event.NotificateBefore,
	)
	return err
}

func (s *Storage) Delete(ctx context.Context, id string) error {
	_, err := s.conn.Exec(ctx, "DELETE FROM calendar_event WHERE id = $1", id)
	return err
}

func (s *Storage) GetByID(ctx context.Context, id string) (*storage.Event, error) {
	rows, err := s.conn.Query(ctx, "SELECT * FROM calendar_event WHERE id = $1", id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	// закрытие строк происходит в CollectRows
	// defer rows.Close()
	event, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[storage.Event])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}

		return nil, err
	}

	return &event, nil
}

func (s *Storage) GetAll(ctx context.Context) ([]storage.Event, error) {
	rows, err := s.conn.Query(ctx, "SELECT * FROM calendar_event ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	// закрытие строк происходит в CollectRows
	// defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[storage.Event])
}

// поиск событий по пересечению с переданным событием.
// пересечение срабатывает только в рамках одного пользователя.
func (s *Storage) GetIntersected(ctx context.Context, event storage.Event) (*storage.Event, error) {
	rows, err := s.conn.Query(
		ctx,
		`
			select *
			from calendar_event
			where
				user_id = $1 AND
				on_date <= $2 AND
				on_date + duration >= $3 AND
				id <> $4
		`,
		event.UserID,
		event.OnDate.Add(event.Duration),
		event.OnDate,
		event.ID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}

		return nil, fmt.Errorf("cant  err: %w", err)
	}

	eventIntersected, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[storage.Event])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}

		return nil, err
	}

	return &eventIntersected, nil
}

func (s *Storage) assertEventValid(ctx context.Context, event storage.Event) error {
	if len(event.ID) == 0 {
		return fmt.Errorf("event id is empty")
	}

	if event.UserID <= 0 {
		return storage.ErrUserIDRequired
	}

	if event.OnDate.IsZero() {
		return storage.ErrOnDateRequired
	}

	eIntersected, err := s.GetIntersected(ctx, event)
	if err != nil && !errors.Is(err, storage.ErrNotFound) {
		return err
	} else if err == nil {
		return fmt.Errorf(
			"err: %w, intersection with event: {ID: %s, OnDate: %s, Duration: %s}",
			storage.ErrDateBusy,
			eIntersected.ID,
			eIntersected.OnDate,
			eIntersected.Duration,
		)
	}

	return nil
}
