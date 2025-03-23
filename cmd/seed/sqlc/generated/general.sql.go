// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: general.sql

package db_seed

import (
	"context"
	"time"

	"api/internal/custom_types"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

const insertCourses = `-- name: InsertCourses :exec
INSERT INTO courses (name, description, capacity)
VALUES (unnest($1::text[]),
        unnest($2::text[]),
        unnest($3::int[]))
RETURNING id
`

type InsertCoursesParams struct {
	NameArray        []string `json:"name_array"`
	DescriptionArray []string `json:"description_array"`
	CapacityArray    []int32  `json:"capacity_array"`
}

func (q *Queries) InsertCourses(ctx context.Context, arg InsertCoursesParams) error {
	_, err := q.db.ExecContext(ctx, insertCourses, pq.Array(arg.NameArray), pq.Array(arg.DescriptionArray), pq.Array(arg.CapacityArray))
	return err
}

const insertCustomersEnrollments = `-- name: InsertCustomersEnrollments :many
WITH prepared_data AS (SELECT unnest($1::uuid[])          AS customer_id,
                              unnest($2::uuid[])             AS event_id,
                              unnest($3::timestamptz[]) AS raw_checked_in_at,
                              unnest($4::bool[])         AS is_cancelled)
INSERT
INTO events.customer_enrollment(customer_id, event_id, checked_in_at, is_cancelled)
SELECT customer_id,
       event_id,
       NULLIF(raw_checked_in_at, '0001-01-01 00:00:00 UTC') AS checked_in_at,
       is_cancelled
FROM prepared_data
RETURNING id
`

type InsertCustomersEnrollmentsParams struct {
	CustomerIDArray  []uuid.UUID `json:"customer_id_array"`
	EventIDArray     []uuid.UUID `json:"event_id_array"`
	CheckedInAtArray []time.Time `json:"checked_in_at_array"`
	IsCancelledArray []bool      `json:"is_cancelled_array"`
}

func (q *Queries) InsertCustomersEnrollments(ctx context.Context, arg InsertCustomersEnrollmentsParams) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, insertCustomersEnrollments,
		pq.Array(arg.CustomerIDArray),
		pq.Array(arg.EventIDArray),
		pq.Array(arg.CheckedInAtArray),
		pq.Array(arg.IsCancelledArray),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertEvents = `-- name: InsertEvents :many
WITH events_data AS (SELECT unnest($1::timestamptz[]) as program_start_at,
                            unnest($2::timestamptz[])   as program_end_at,
                            unnest($3::timetz[]) AS event_start_time,
                            unnest($4::timetz[])   AS event_end_time,
                            unnest($5::day_enum[])                 AS day,
                            unnest($6::text[])           AS practice_name,
                            unnest($7::text[])             AS course_name,
                            unnest($8::text[])               AS game_name,
                            unnest($9::text[])           as location_name)
INSERT
INTO events.events (program_start_at, program_end_at, event_start_time, event_end_time, day, practice_id, course_id,
                    game_id, location_id)
SELECT e.program_start_at,
       e.program_end_at,
       e.event_start_time,
       e.event_end_time,
       e.day,
       p.id AS practice_id,
       c.id AS course_id,
       g.id AS game_id,
       l.id AS location_id
FROM events_data e
         LEFT JOIN LATERAL (SELECT id FROM public.practices WHERE name = e.practice_name) p ON TRUE
         LEFT JOIN LATERAL (SELECT id FROM courses WHERE name = e.course_name) c ON TRUE
         LEFT JOIN LATERAL (SELECT id FROM public.games WHERE name = e.game_name) g ON TRUE
         LEFT JOIN LATERAL (SELECT id FROM location.locations WHERE name = e.location_name) l ON TRUE
RETURNING id
`

type InsertEventsParams struct {
	ProgramStartAtArray []time.Time                     `json:"program_start_at_array"`
	ProgramEndAtArray   []time.Time                     `json:"program_end_at_array"`
	EventStartTimeArray []custom_types.TimeWithTimeZone `json:"event_start_time_array"`
	EventEndTimeArray   []custom_types.TimeWithTimeZone `json:"event_end_time_array"`
	DayArray            []DayEnum                       `json:"day_array"`
	PracticeNameArray   []string                        `json:"practice_name_array"`
	CourseNameArray     []string                        `json:"course_name_array"`
	GameNameArray       []string                        `json:"game_name_array"`
	LocationNameArray   []string                        `json:"location_name_array"`
}

func (q *Queries) InsertEvents(ctx context.Context, arg InsertEventsParams) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, insertEvents,
		pq.Array(arg.ProgramStartAtArray),
		pq.Array(arg.ProgramEndAtArray),
		pq.Array(arg.EventStartTimeArray),
		pq.Array(arg.EventEndTimeArray),
		pq.Array(arg.DayArray),
		pq.Array(arg.PracticeNameArray),
		pq.Array(arg.CourseNameArray),
		pq.Array(arg.GameNameArray),
		pq.Array(arg.LocationNameArray),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertGames = `-- name: InsertGames :exec
INSERT INTO games (name)
VALUES (unnest($1::text[]))
RETURNING id
`

func (q *Queries) InsertGames(ctx context.Context, nameArray []string) error {
	_, err := q.db.ExecContext(ctx, insertGames, pq.Array(nameArray))
	return err
}

const insertLocations = `-- name: InsertLocations :exec
INSERT INTO location.locations (name, address)
VALUES (unnest($1::text[]), unnest($2::text[]))
RETURNING id
`

type InsertLocationsParams struct {
	NameArray    []string `json:"name_array"`
	AddressArray []string `json:"address_array"`
}

func (q *Queries) InsertLocations(ctx context.Context, arg InsertLocationsParams) error {
	_, err := q.db.ExecContext(ctx, insertLocations, pq.Array(arg.NameArray), pq.Array(arg.AddressArray))
	return err
}

const insertPractices = `-- name: InsertPractices :exec
INSERT INTO practices (name, description, level, capacity)
VALUES (unnest($1::text[]),
        unnest($2::text[]),
        unnest($3::practice_level[]),
        unnest($4::int[]))
RETURNING id
`

type InsertPracticesParams struct {
	NameArray        []string        `json:"name_array"`
	DescriptionArray []string        `json:"description_array"`
	LevelArray       []PracticeLevel `json:"level_array"`
	CapacityArray    []int32         `json:"capacity_array"`
}

func (q *Queries) InsertPractices(ctx context.Context, arg InsertPracticesParams) error {
	_, err := q.db.ExecContext(ctx, insertPractices,
		pq.Array(arg.NameArray),
		pq.Array(arg.DescriptionArray),
		pq.Array(arg.LevelArray),
		pq.Array(arg.CapacityArray),
	)
	return err
}
