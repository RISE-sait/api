// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: event_queries.sql

package event_db

import (
	"context"
	"database/sql"
	"time"

	"api/internal/custom_types"
	"github.com/google/uuid"
)

const createEvent = `-- name: CreateEvent :exec
INSERT INTO events (program_start_at, program_end_at, event_start_time, event_end_time, day, location_id, course_id,
                    practice_id, game_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

type CreateEventParams struct {
	ProgramStartAt time.Time                     `json:"program_start_at"`
	ProgramEndAt   time.Time                     `json:"program_end_at"`
	EventStartTime custom_types.TimeWithTimeZone `json:"event_start_time"`
	EventEndTime   custom_types.TimeWithTimeZone `json:"event_end_time"`
	Day            DayEnum                       `json:"day"`
	LocationID     uuid.NullUUID                 `json:"location_id"`
	CourseID       uuid.NullUUID                 `json:"course_id"`
	PracticeID     uuid.NullUUID                 `json:"practice_id"`
	GameID         uuid.NullUUID                 `json:"game_id"`
}

func (q *Queries) CreateEvent(ctx context.Context, arg CreateEventParams) error {
	_, err := q.db.ExecContext(ctx, createEvent,
		arg.ProgramStartAt,
		arg.ProgramEndAt,
		arg.EventStartTime,
		arg.EventEndTime,
		arg.Day,
		arg.LocationID,
		arg.CourseID,
		arg.PracticeID,
		arg.GameID,
	)
	return err
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
SELECT e.id, e.program_start_at, e.program_end_at, e.practice_id, e.course_id, e.game_id, e.location_id, e.created_at, e.updated_at, e.day, e.event_start_time, e.event_end_time,
       p.name as practice_name,
       p.description as practice_description,
       c.name as course_name,
       c.description as course_description,
       g.name as game_name,
       l.name as location_name,
       l.address as address
FROM public.events e
         LEFT JOIN public.practices p ON e.practice_id = p.id
         LEFT JOIN course.courses c ON e.course_id = c.id
         LEFT JOIN public.games g ON e.game_id = g.id
         LEFT JOIN location.locations l ON e.location_id = l.id
WHERE e.id = $1
`

type GetEventByIdRow struct {
	ID                  uuid.UUID                     `json:"id"`
	ProgramStartAt      time.Time                     `json:"program_start_at"`
	ProgramEndAt        time.Time                     `json:"program_end_at"`
	PracticeID          uuid.NullUUID                 `json:"practice_id"`
	CourseID            uuid.NullUUID                 `json:"course_id"`
	GameID              uuid.NullUUID                 `json:"game_id"`
	LocationID          uuid.NullUUID                 `json:"location_id"`
	CreatedAt           time.Time                     `json:"created_at"`
	UpdatedAt           time.Time                     `json:"updated_at"`
	Day                 DayEnum                       `json:"day"`
	EventStartTime      custom_types.TimeWithTimeZone `json:"event_start_time"`
	EventEndTime        custom_types.TimeWithTimeZone `json:"event_end_time"`
	PracticeName        sql.NullString                `json:"practice_name"`
	PracticeDescription sql.NullString                `json:"practice_description"`
	CourseName          sql.NullString                `json:"course_name"`
	CourseDescription   sql.NullString                `json:"course_description"`
	GameName            sql.NullString                `json:"game_name"`
	LocationName        sql.NullString                `json:"location_name"`
	Address             sql.NullString                `json:"address"`
}

func (q *Queries) GetEventById(ctx context.Context, id uuid.UUID) (GetEventByIdRow, error) {
	row := q.db.QueryRowContext(ctx, getEventById, id)
	var i GetEventByIdRow
	err := row.Scan(
		&i.ID,
		&i.ProgramStartAt,
		&i.ProgramEndAt,
		&i.PracticeID,
		&i.CourseID,
		&i.GameID,
		&i.LocationID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Day,
		&i.EventStartTime,
		&i.EventEndTime,
		&i.PracticeName,
		&i.PracticeDescription,
		&i.CourseName,
		&i.CourseDescription,
		&i.GameName,
		&i.LocationName,
		&i.Address,
	)
	return i, err
}

const getEvents = `-- name: GetEvents :many
SELECT e.id, e.program_start_at, e.program_end_at, e.practice_id, e.course_id, e.game_id, e.location_id, e.created_at, e.updated_at, e.day, e.event_start_time, e.event_end_time,
       p.name as practice_name,
       p.description as practice_description,
       c.name as course_name,
       c.description as course_description,
       g.name as game_name,
       l.name as location_name,
       l.address as address
FROM public.events e
LEFT JOIN public.practices p ON e.practice_id = p.id
LEFT JOIN course.courses c ON e.course_id = c.id
LEFT JOIN public.games g ON e.game_id = g.id
LEFT JOIN location.locations l ON e.location_id = l.id
WHERE ($1 = course_id OR $1 IS NULL)
  AND ($2 = game_id OR $2 IS NULL)
  AND ($3 = practice_id OR $3 IS NULL)
  AND ($4 = location_id OR $4 IS NULL)
  AND ($5 >= e.program_start_at OR $5 IS NULL) -- within boundary
  AND ($6 <= e.program_end_at OR $6 IS NULL)
`

type GetEventsParams struct {
	CourseID   uuid.NullUUID `json:"course_id"`
	GameID     uuid.NullUUID `json:"game_id"`
	PracticeID uuid.NullUUID `json:"practice_id"`
	LocationID uuid.NullUUID `json:"location_id"`
	Before     sql.NullTime  `json:"before"`
	After      sql.NullTime  `json:"after"`
}

type GetEventsRow struct {
	ID                  uuid.UUID                     `json:"id"`
	ProgramStartAt      time.Time                     `json:"program_start_at"`
	ProgramEndAt        time.Time                     `json:"program_end_at"`
	PracticeID          uuid.NullUUID                 `json:"practice_id"`
	CourseID            uuid.NullUUID                 `json:"course_id"`
	GameID              uuid.NullUUID                 `json:"game_id"`
	LocationID          uuid.NullUUID                 `json:"location_id"`
	CreatedAt           time.Time                     `json:"created_at"`
	UpdatedAt           time.Time                     `json:"updated_at"`
	Day                 DayEnum                       `json:"day"`
	EventStartTime      custom_types.TimeWithTimeZone `json:"event_start_time"`
	EventEndTime        custom_types.TimeWithTimeZone `json:"event_end_time"`
	PracticeName        sql.NullString                `json:"practice_name"`
	PracticeDescription sql.NullString                `json:"practice_description"`
	CourseName          sql.NullString                `json:"course_name"`
	CourseDescription   sql.NullString                `json:"course_description"`
	GameName            sql.NullString                `json:"game_name"`
	LocationName        sql.NullString                `json:"location_name"`
	Address             sql.NullString                `json:"address"`
}

func (q *Queries) GetEvents(ctx context.Context, arg GetEventsParams) ([]GetEventsRow, error) {
	rows, err := q.db.QueryContext(ctx, getEvents,
		arg.CourseID,
		arg.GameID,
		arg.PracticeID,
		arg.LocationID,
		arg.Before,
		arg.After,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetEventsRow
	for rows.Next() {
		var i GetEventsRow
		if err := rows.Scan(
			&i.ID,
			&i.ProgramStartAt,
			&i.ProgramEndAt,
			&i.PracticeID,
			&i.CourseID,
			&i.GameID,
			&i.LocationID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Day,
			&i.EventStartTime,
			&i.EventEndTime,
			&i.PracticeName,
			&i.PracticeDescription,
			&i.CourseName,
			&i.CourseDescription,
			&i.GameName,
			&i.LocationName,
			&i.Address,
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

const updateEvent = `-- name: UpdateEvent :exec
UPDATE events
SET program_start_at   = $1,
    program_end_at     = $2,
    location_id    = $3,
    practice_id    = $4,
    course_id      = $5,
    game_id        = $6,
    event_start_time = $7,
    event_end_time   = $8,
    day                = $9,
    updated_at     = current_timestamp
WHERE id = $10
`

type UpdateEventParams struct {
	ProgramStartAt time.Time                     `json:"program_start_at"`
	ProgramEndAt   time.Time                     `json:"program_end_at"`
	LocationID     uuid.NullUUID                 `json:"location_id"`
	PracticeID     uuid.NullUUID                 `json:"practice_id"`
	CourseID       uuid.NullUUID                 `json:"course_id"`
	GameID         uuid.NullUUID                 `json:"game_id"`
	EventStartTime custom_types.TimeWithTimeZone `json:"event_start_time"`
	EventEndTime   custom_types.TimeWithTimeZone `json:"event_end_time"`
	Day            DayEnum                       `json:"day"`
	ID             uuid.UUID                     `json:"id"`
}

func (q *Queries) UpdateEvent(ctx context.Context, arg UpdateEventParams) error {
	_, err := q.db.ExecContext(ctx, updateEvent,
		arg.ProgramStartAt,
		arg.ProgramEndAt,
		arg.LocationID,
		arg.PracticeID,
		arg.CourseID,
		arg.GameID,
		arg.EventStartTime,
		arg.EventEndTime,
		arg.Day,
		arg.ID,
	)
	return err
}
