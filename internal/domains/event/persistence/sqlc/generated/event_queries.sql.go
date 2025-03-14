// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: event_queries.sql

package event_db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createEvent = `-- name: CreateEvent :one
INSERT INTO events (event_start_at, event_end_at, location_id, course_id, practice_id, game_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, event_start_at, event_end_at, practice_id, course_id, game_id, location_id, created_at, updated_at
`

type CreateEventParams struct {
	EventStartAt time.Time     `json:"event_start_at"`
	EventEndAt   time.Time     `json:"event_end_at"`
	LocationID   uuid.UUID     `json:"location_id"`
	CourseID     uuid.NullUUID `json:"course_id"`
	PracticeID   uuid.NullUUID `json:"practice_id"`
	GameID       uuid.NullUUID `json:"game_id"`
}

func (q *Queries) CreateEvent(ctx context.Context, arg CreateEventParams) (Event, error) {
	row := q.db.QueryRowContext(ctx, createEvent,
		arg.EventStartAt,
		arg.EventEndAt,
		arg.LocationID,
		arg.CourseID,
		arg.PracticeID,
		arg.GameID,
	)
	var i Event
	err := row.Scan(
		&i.ID,
		&i.EventStartAt,
		&i.EventEndAt,
		&i.PracticeID,
		&i.CourseID,
		&i.GameID,
		&i.LocationID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteEvent = `-- name: DeleteEvent :exec
DELETE
FROM events
WHERE id = $1
`

func (q *Queries) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, deleteEvent, id)
	return err
}

const getEventById = `-- name: GetEventById :one
SELECT id, event_start_at, event_end_at, practice_id, course_id, game_id, location_id, created_at, updated_at
FROM events
WHERE id = $1
`

func (q *Queries) GetEventById(ctx context.Context, id uuid.UUID) (Event, error) {
	row := q.db.QueryRowContext(ctx, getEventById, id)
	var i Event
	err := row.Scan(
		&i.ID,
		&i.EventStartAt,
		&i.EventEndAt,
		&i.PracticeID,
		&i.CourseID,
		&i.GameID,
		&i.LocationID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getEvents = `-- name: GetEvents :many
SELECT id, event_start_at, event_end_at, practice_id, course_id, game_id, location_id, created_at, updated_at
FROM events
WHERE event_start_at >= $1
  AND event_end_at <= $2
  AND ($3 = course_id OR $3 IS NULL)
  AND ($4 = game_id OR $4 IS NULL)
    AND ($5 = practice_id OR $5 IS NULL)
        AND ($6 = location_id OR $6 IS NULL)
`

type GetEventsParams struct {
	After      time.Time     `json:"after"`
	Before     time.Time     `json:"before"`
	CourseID   uuid.NullUUID `json:"course_id"`
	GameID     uuid.NullUUID `json:"game_id"`
	PracticeID uuid.NullUUID `json:"practice_id"`
	LocationID uuid.NullUUID `json:"location_id"`
}

func (q *Queries) GetEvents(ctx context.Context, arg GetEventsParams) ([]Event, error) {
	rows, err := q.db.QueryContext(ctx, getEvents,
		arg.After,
		arg.Before,
		arg.CourseID,
		arg.GameID,
		arg.PracticeID,
		arg.LocationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Event
	for rows.Next() {
		var i Event
		if err := rows.Scan(
			&i.ID,
			&i.EventStartAt,
			&i.EventEndAt,
			&i.PracticeID,
			&i.CourseID,
			&i.GameID,
			&i.LocationID,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateEvent = `-- name: UpdateEvent :one
UPDATE events
SET event_start_at = $1,
    event_end_at   = $2,
    location_id    = $3,
    practice_id    = $4,
    course_id      = $5,
    game_id        = $6,
    updated_at     = current_timestamp
WHERE id = $7
RETURNING id, event_start_at, event_end_at, practice_id, course_id, game_id, location_id, created_at, updated_at
`

type UpdateEventParams struct {
	EventStartAt time.Time     `json:"event_start_at"`
	EventEndAt   time.Time     `json:"event_end_at"`
	LocationID   uuid.UUID     `json:"location_id"`
	PracticeID   uuid.NullUUID `json:"practice_id"`
	CourseID     uuid.NullUUID `json:"course_id"`
	GameID       uuid.NullUUID `json:"game_id"`
	ID           uuid.UUID     `json:"id"`
}

func (q *Queries) UpdateEvent(ctx context.Context, arg UpdateEventParams) (Event, error) {
	row := q.db.QueryRowContext(ctx, updateEvent,
		arg.EventStartAt,
		arg.EventEndAt,
		arg.LocationID,
		arg.PracticeID,
		arg.CourseID,
		arg.GameID,
		arg.ID,
	)
	var i Event
	err := row.Scan(
		&i.ID,
		&i.EventStartAt,
		&i.EventEndAt,
		&i.PracticeID,
		&i.CourseID,
		&i.GameID,
		&i.LocationID,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
