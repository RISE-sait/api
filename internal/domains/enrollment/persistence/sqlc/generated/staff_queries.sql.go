// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: staff_queries.sql

package db_event

import (
	"context"

	"github.com/google/uuid"
)

const assignStaffToEvent = `-- name: AssignStaffToEvent :execrows
INSERT INTO events.staff (event_id, staff_id)
VALUES ($1, $2)
`

type AssignStaffToEventParams struct {
	EventID uuid.UUID `json:"event_id"`
	StaffID uuid.UUID `json:"staff_id"`
}

func (q *Queries) AssignStaffToEvent(ctx context.Context, arg AssignStaffToEventParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, assignStaffToEvent, arg.EventID, arg.StaffID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const unassignStaffFromEvent = `-- name: UnassignStaffFromEvent :execrows
DELETE
FROM events.staff
where staff_id = $1
  and event_id = $2
`

type UnassignStaffFromEventParams struct {
	StaffID uuid.UUID `json:"staff_id"`
	EventID uuid.UUID `json:"event_id"`
}

func (q *Queries) UnassignStaffFromEvent(ctx context.Context, arg UnassignStaffFromEventParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, unassignStaffFromEvent, arg.StaffID, arg.EventID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
