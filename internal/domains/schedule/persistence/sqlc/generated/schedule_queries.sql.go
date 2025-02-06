// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: schedule_queries.sql

package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createSchedule = `-- name: CreateSchedule :execrows
INSERT INTO schedules (begin_time, end_time, facility_id, course_id, day)
VALUES ($1, $2, $3, $4, $5)
`

type CreateScheduleParams struct {
	BeginTime  time.Time     `json:"begin_time"`
	EndTime    time.Time     `json:"end_time"`
	FacilityID uuid.UUID     `json:"facility_id"`
	CourseID   uuid.NullUUID `json:"course_id"`
	Day        DayEnum       `json:"day"`
}

func (q *Queries) CreateSchedule(ctx context.Context, arg CreateScheduleParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, createSchedule,
		arg.BeginTime,
		arg.EndTime,
		arg.FacilityID,
		arg.CourseID,
		arg.Day,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const deleteSchedule = `-- name: DeleteSchedule :execrows
DELETE FROM schedules WHERE id = $1
`

func (q *Queries) DeleteSchedule(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteSchedule, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getSchedules = `-- name: GetSchedules :many
SELECT s.id, begin_time, end_time, s.day, c.name as course, f.name as facility FROM schedules s
JOIN courses c ON c.id = s.course_id
JOIN facilities f ON f.id = s.facility_id
WHERE 
    (begin_time >= $1 OR $1::text LIKE '%00:00:00%')
    AND (end_time <= $2 OR $2::text LIKE '%00:00:00%')
   AND (facility_id = $3 OR $3 = '00000000-0000-0000-0000-000000000000')
    AND (course_id = $4 or $4 IS NULL)
`

type GetSchedulesParams struct {
	BeginTime  time.Time     `json:"begin_time"`
	EndTime    time.Time     `json:"end_time"`
	FacilityID uuid.UUID     `json:"facility_id"`
	CourseID   uuid.NullUUID `json:"course_id"`
}

type GetSchedulesRow struct {
	ID        uuid.UUID `json:"id"`
	BeginTime time.Time `json:"begin_time"`
	EndTime   time.Time `json:"end_time"`
	Day       DayEnum   `json:"day"`
	Course    string    `json:"course"`
	Facility  string    `json:"facility"`
}

func (q *Queries) GetSchedules(ctx context.Context, arg GetSchedulesParams) ([]GetSchedulesRow, error) {
	rows, err := q.db.QueryContext(ctx, getSchedules,
		arg.BeginTime,
		arg.EndTime,
		arg.FacilityID,
		arg.CourseID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetSchedulesRow
	for rows.Next() {
		var i GetSchedulesRow
		if err := rows.Scan(
			&i.ID,
			&i.BeginTime,
			&i.EndTime,
			&i.Day,
			&i.Course,
			&i.Facility,
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

const updateSchedule = `-- name: UpdateSchedule :execrows
UPDATE schedules s
SET begin_time = $1, end_time = $2, facility_id = $3, course_id = $4, day = $5
WHERE s.id = $6
`

type UpdateScheduleParams struct {
	BeginTime  time.Time     `json:"begin_time"`
	EndTime    time.Time     `json:"end_time"`
	FacilityID uuid.UUID     `json:"facility_id"`
	CourseID   uuid.NullUUID `json:"course_id"`
	Day        DayEnum       `json:"day"`
	ID         uuid.UUID     `json:"id"`
}

func (q *Queries) UpdateSchedule(ctx context.Context, arg UpdateScheduleParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, updateSchedule,
		arg.BeginTime,
		arg.EndTime,
		arg.FacilityID,
		arg.CourseID,
		arg.Day,
		arg.ID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
