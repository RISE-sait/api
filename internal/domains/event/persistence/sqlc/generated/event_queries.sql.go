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
INSERT INTO events.events (program_start_at, program_end_at, event_start_time, event_end_time, day, location_id,
                           program_id, capacity, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5,
        $6, $7, $8, $9::uuid, $9::uuid)
`

type CreateEventParams struct {
	ProgramStartAt time.Time                     `json:"program_start_at"`
	ProgramEndAt   time.Time                     `json:"program_end_at"`
	EventStartTime custom_types.TimeWithTimeZone `json:"event_start_time"`
	EventEndTime   custom_types.TimeWithTimeZone `json:"event_end_time"`
	Day            DayEnum                       `json:"day"`
	LocationID     uuid.UUID                     `json:"location_id"`
	ProgramID      uuid.NullUUID                 `json:"program_id"`
	Capacity       sql.NullInt32                 `json:"capacity"`
	CreatedBy      uuid.UUID                     `json:"created_by"`
}

func (q *Queries) CreateEvent(ctx context.Context, arg CreateEventParams) error {
	_, err := q.db.ExecContext(ctx, createEvent,
		arg.ProgramStartAt,
		arg.ProgramEndAt,
		arg.EventStartTime,
		arg.EventEndTime,
		arg.Day,
		arg.LocationID,
		arg.ProgramID,
		arg.Capacity,
		arg.CreatedBy,
	)
	return err
}

const deleteEvent = `-- name: DeleteEvent :exec
DELETE
FROM events.events
WHERE id = $1
`

func (q *Queries) DeleteEvent(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, deleteEvent, id)
	return err
}

const getEventById = `-- name: GetEventById :many
SELECT
    e.id, e.program_start_at, e.program_end_at, e.program_id, e.team_id, e.location_id, e.capacity, e.created_at, e.updated_at, e.day, e.event_start_time, e.event_end_time, e.created_by, e.updated_by,
    p.name AS program_name,
    p.description AS program_description,
    p."type" AS program_type,
    l.name AS location_name,
    l.address AS location_address,
    -- Staff fields
    s.id AS staff_id,
    sr.role_name AS staff_role_name,
    us.email AS staff_email,
    us.first_name AS staff_first_name,
    us.last_name AS staff_last_name,
    us.gender AS staff_gender,
    us.phone AS staff_phone,
    -- Customer fields
    uc.id AS customer_id,
    uc.first_name AS customer_first_name,
    uc.last_name AS customer_last_name,
    uc.email AS customer_email,
    uc.phone AS customer_phone,
    uc.gender AS customer_gender,
    ce.is_cancelled AS customer_is_cancelled,
    -- Team field (added missing team reference)
    t.id AS team_id,
    t.name AS team_name
FROM events.events e
         LEFT JOIN program.programs p ON e.program_id = p.id
         LEFT JOIN location.locations l ON e.location_id = l.id
         LEFT JOIN events.staff es ON e.id = es.event_id
         LEFT JOIN staff.staff s ON es.staff_id = s.id
         LEFT JOIN staff.staff_roles sr ON s.role_id = sr.id
         LEFT JOIN users.users us ON s.id = us.id
         LEFT JOIN events.customer_enrollment ce ON e.id = ce.event_id
         LEFT JOIN users.users uc ON ce.customer_id = uc.id
         LEFT JOIN athletic.teams t ON t.id = e.team_id
WHERE e.id = $1
ORDER BY s.id, uc.id
`

type GetEventByIdRow struct {
	ID                  uuid.UUID                     `json:"id"`
	ProgramStartAt      time.Time                     `json:"program_start_at"`
	ProgramEndAt        time.Time                     `json:"program_end_at"`
	ProgramID           uuid.NullUUID                 `json:"program_id"`
	TeamID              uuid.NullUUID                 `json:"team_id"`
	LocationID          uuid.UUID                     `json:"location_id"`
	Capacity            sql.NullInt32                 `json:"capacity"`
	CreatedAt           time.Time                     `json:"created_at"`
	UpdatedAt           time.Time                     `json:"updated_at"`
	Day                 DayEnum                       `json:"day"`
	EventStartTime      custom_types.TimeWithTimeZone `json:"event_start_time"`
	EventEndTime        custom_types.TimeWithTimeZone `json:"event_end_time"`
	CreatedBy           uuid.NullUUID                 `json:"created_by"`
	UpdatedBy           uuid.NullUUID                 `json:"updated_by"`
	ProgramName         sql.NullString                `json:"program_name"`
	ProgramDescription  sql.NullString                `json:"program_description"`
	ProgramType         NullProgramProgramType        `json:"program_type"`
	LocationName        sql.NullString                `json:"location_name"`
	LocationAddress     sql.NullString                `json:"location_address"`
	StaffID             uuid.NullUUID                 `json:"staff_id"`
	StaffRoleName       sql.NullString                `json:"staff_role_name"`
	StaffEmail          sql.NullString                `json:"staff_email"`
	StaffFirstName      sql.NullString                `json:"staff_first_name"`
	StaffLastName       sql.NullString                `json:"staff_last_name"`
	StaffGender         sql.NullString                `json:"staff_gender"`
	StaffPhone          sql.NullString                `json:"staff_phone"`
	CustomerID          uuid.NullUUID                 `json:"customer_id"`
	CustomerFirstName   sql.NullString                `json:"customer_first_name"`
	CustomerLastName    sql.NullString                `json:"customer_last_name"`
	CustomerEmail       sql.NullString                `json:"customer_email"`
	CustomerPhone       sql.NullString                `json:"customer_phone"`
	CustomerGender      sql.NullString                `json:"customer_gender"`
	CustomerIsCancelled sql.NullBool                  `json:"customer_is_cancelled"`
	TeamID_2            uuid.NullUUID                 `json:"team_id_2"`
	TeamName            sql.NullString                `json:"team_name"`
}

func (q *Queries) GetEventById(ctx context.Context, id uuid.UUID) ([]GetEventByIdRow, error) {
	rows, err := q.db.QueryContext(ctx, getEventById, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetEventByIdRow
	for rows.Next() {
		var i GetEventByIdRow
		if err := rows.Scan(
			&i.ID,
			&i.ProgramStartAt,
			&i.ProgramEndAt,
			&i.ProgramID,
			&i.TeamID,
			&i.LocationID,
			&i.Capacity,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Day,
			&i.EventStartTime,
			&i.EventEndTime,
			&i.CreatedBy,
			&i.UpdatedBy,
			&i.ProgramName,
			&i.ProgramDescription,
			&i.ProgramType,
			&i.LocationName,
			&i.LocationAddress,
			&i.StaffID,
			&i.StaffRoleName,
			&i.StaffEmail,
			&i.StaffFirstName,
			&i.StaffLastName,
			&i.StaffGender,
			&i.StaffPhone,
			&i.CustomerID,
			&i.CustomerFirstName,
			&i.CustomerLastName,
			&i.CustomerEmail,
			&i.CustomerPhone,
			&i.CustomerGender,
			&i.CustomerIsCancelled,
			&i.TeamID_2,
			&i.TeamName,
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

const getEvents = `-- name: GetEvents :many
SELECT DISTINCT e.id, e.program_start_at, e.program_end_at, e.program_id, e.team_id, e.location_id, e.capacity, e.created_at, e.updated_at, e.day, e.event_start_time, e.event_end_time, e.created_by, e.updated_by,
                p.name        AS program_name,
                p.description AS program_description,
                p."type"      AS program_type,
                l.name        AS location_name,
                l.address     AS location_address,
                t.name        as team_name
FROM events.events e
         LEFT JOIN program.programs p ON e.program_id = p.id
         JOIN location.locations l ON e.location_id = l.id
         LEFT JOIN events.staff es ON e.id = es.event_id
         LEFT JOIN events.customer_enrollment ce ON e.id = ce.event_id
         LEFT JOIN athletic.teams t ON t.id = e.team_id
WHERE (
          ($1::uuid = e.program_id OR $1 IS NULL)
              AND ($2::uuid = e.location_id OR $2 IS NULL)
              AND ($3::timestamp >= e.program_start_at OR $3 IS NULL)
              AND ($4::timestamp <= e.program_end_at OR $4 IS NULL)
              AND ($5 = p.type OR $5 IS NULL)
              AND ($6::uuid IS NULL OR ce.customer_id = $6::uuid OR
                   es.staff_id = $6::uuid)
              AND ($7::uuid IS NULL OR e.team_id = $7)
                AND ($8::uuid IS NULL OR e.created_by = $8)
                AND ($9::uuid IS NULL OR e.updated_by = $9)
          )
`

type GetEventsParams struct {
	ProgramID  uuid.NullUUID          `json:"program_id"`
	LocationID uuid.NullUUID          `json:"location_id"`
	Before     sql.NullTime           `json:"before"`
	After      sql.NullTime           `json:"after"`
	Type       NullProgramProgramType `json:"type"`
	UserID     uuid.NullUUID          `json:"user_id"`
	TeamID     uuid.NullUUID          `json:"team_id"`
	CreatedBy  uuid.NullUUID          `json:"created_by"`
	UpdatedBy  uuid.NullUUID          `json:"updated_by"`
}

type GetEventsRow struct {
	ID                 uuid.UUID                     `json:"id"`
	ProgramStartAt     time.Time                     `json:"program_start_at"`
	ProgramEndAt       time.Time                     `json:"program_end_at"`
	ProgramID          uuid.NullUUID                 `json:"program_id"`
	TeamID             uuid.NullUUID                 `json:"team_id"`
	LocationID         uuid.UUID                     `json:"location_id"`
	Capacity           sql.NullInt32                 `json:"capacity"`
	CreatedAt          time.Time                     `json:"created_at"`
	UpdatedAt          time.Time                     `json:"updated_at"`
	Day                DayEnum                       `json:"day"`
	EventStartTime     custom_types.TimeWithTimeZone `json:"event_start_time"`
	EventEndTime       custom_types.TimeWithTimeZone `json:"event_end_time"`
	CreatedBy          uuid.NullUUID                 `json:"created_by"`
	UpdatedBy          uuid.NullUUID                 `json:"updated_by"`
	ProgramName        sql.NullString                `json:"program_name"`
	ProgramDescription sql.NullString                `json:"program_description"`
	ProgramType        NullProgramProgramType        `json:"program_type"`
	LocationName       string                        `json:"location_name"`
	LocationAddress    string                        `json:"location_address"`
	TeamName           sql.NullString                `json:"team_name"`
}

func (q *Queries) GetEvents(ctx context.Context, arg GetEventsParams) ([]GetEventsRow, error) {
	rows, err := q.db.QueryContext(ctx, getEvents,
		arg.ProgramID,
		arg.LocationID,
		arg.Before,
		arg.After,
		arg.Type,
		arg.UserID,
		arg.TeamID,
		arg.CreatedBy,
		arg.UpdatedBy,
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
			&i.ProgramID,
			&i.TeamID,
			&i.LocationID,
			&i.Capacity,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Day,
			&i.EventStartTime,
			&i.EventEndTime,
			&i.CreatedBy,
			&i.UpdatedBy,
			&i.ProgramName,
			&i.ProgramDescription,
			&i.ProgramType,
			&i.LocationName,
			&i.LocationAddress,
			&i.TeamName,
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
UPDATE events.events
SET program_start_at = $1,
    program_end_at   = $2,
    location_id      = $3,
    program_id       = $4,
    event_start_time = $5,
    event_end_time   = $6,
    day              = $7,
    capacity         = $8,
    updated_at       = current_timestamp,
    updated_by = $10::uuid
WHERE id = $9
`

type UpdateEventParams struct {
	ProgramStartAt time.Time                     `json:"program_start_at"`
	ProgramEndAt   time.Time                     `json:"program_end_at"`
	LocationID     uuid.UUID                     `json:"location_id"`
	ProgramID      uuid.NullUUID                 `json:"program_id"`
	EventStartTime custom_types.TimeWithTimeZone `json:"event_start_time"`
	EventEndTime   custom_types.TimeWithTimeZone `json:"event_end_time"`
	Day            DayEnum                       `json:"day"`
	Capacity       sql.NullInt32                 `json:"capacity"`
	ID             uuid.UUID                     `json:"id"`
	UpdatedBy      uuid.UUID                     `json:"updated_by"`
}

func (q *Queries) UpdateEvent(ctx context.Context, arg UpdateEventParams) error {
	_, err := q.db.ExecContext(ctx, updateEvent,
		arg.ProgramStartAt,
		arg.ProgramEndAt,
		arg.LocationID,
		arg.ProgramID,
		arg.EventStartTime,
		arg.EventEndTime,
		arg.Day,
		arg.Capacity,
		arg.ID,
		arg.UpdatedBy,
	)
	return err
}
