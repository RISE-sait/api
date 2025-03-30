// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: event.sql

package db_seed

import (
	"context"
	"time"

	"api/internal/custom_types"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

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
WITH events_data AS (SELECT unnest($1::timestamptz[]) as recurrence_start_at,
                            unnest($2::timestamptz[])   as recurrence_end_at,
                            unnest($3::timetz[])      AS event_start_time,
                            unnest($4::timetz[])        AS event_end_time,
                            unnest($5::day_enum[])                 AS day,
                            unnest($6::text[])           AS program_name,
                            unnest($7::text[])           as location_name)
INSERT
INTO events.events (recurrence_start_at, recurrence_end_at, event_start_time, event_end_time, program_id, day, location_id)
SELECT e.recurrence_start_at,
       NULLIF(e.recurrence_end_at, '0001-01-01 00:00:00 UTC') AS program_end_at,
       e.event_start_time,
       e.event_end_time,
         p.id AS program_id,
       e.day,
       l.id AS location_id
FROM events_data e
         LEFT JOIN LATERAL (SELECT id FROM program.programs WHERE name = e.program_name) p ON TRUE
         LEFT JOIN LATERAL (SELECT id FROM location.locations WHERE name = e.location_name) l ON TRUE
RETURNING id
`

type InsertEventsParams struct {
	RecurringStartAtArray []time.Time                     `json:"recurring_start_at_array"`
	RecurringEndAtArray   []time.Time                     `json:"recurring_end_at_array"`
	EventStartTimeArray   []custom_types.TimeWithTimeZone `json:"event_start_time_array"`
	EventEndTimeArray     []custom_types.TimeWithTimeZone `json:"event_end_time_array"`
	DayArray              []DayEnum                       `json:"day_array"`
	ProgramNameArray      []string                        `json:"program_name_array"`
	LocationNameArray     []string                        `json:"location_name_array"`
}

func (q *Queries) InsertEvents(ctx context.Context, arg InsertEventsParams) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, insertEvents,
		pq.Array(arg.RecurringStartAtArray),
		pq.Array(arg.RecurringEndAtArray),
		pq.Array(arg.EventStartTimeArray),
		pq.Array(arg.EventEndTimeArray),
		pq.Array(arg.DayArray),
		pq.Array(arg.ProgramNameArray),
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

const insertEventsStaff = `-- name: InsertEventsStaff :exec
WITH prepared_data AS (SELECT unnest($1::uuid[]) AS event_id,
                              unnest($2::uuid[]) AS staff_id)
INSERT
INTO events.staff(event_id, staff_id)
SELECT event_id,
       staff_id
FROM prepared_data
`

type InsertEventsStaffParams struct {
	EventIDArray []uuid.UUID `json:"event_id_array"`
	StaffIDArray []uuid.UUID `json:"staff_id_array"`
}

func (q *Queries) InsertEventsStaff(ctx context.Context, arg InsertEventsStaffParams) error {
	_, err := q.db.ExecContext(ctx, insertEventsStaff, pq.Array(arg.EventIDArray), pq.Array(arg.StaffIDArray))
	return err
}
