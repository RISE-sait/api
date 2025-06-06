// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: enrollment_queries.sql

package db_event

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const enrollCustomer = `-- name: EnrollCustomer :one
INSERT INTO events.customer_enrollment (customer_id, event_id, checked_in_at, is_cancelled)
VALUES ($1, $2, $3, false)
RETURNING id, customer_id, event_id, created_at, updated_at, checked_in_at, is_cancelled, payment_status, payment_expired_at
`

type EnrollCustomerParams struct {
	CustomerID  uuid.UUID    `json:"customer_id"`
	EventID     uuid.UUID    `json:"event_id"`
	CheckedInAt sql.NullTime `json:"checked_in_at"`
}

func (q *Queries) EnrollCustomer(ctx context.Context, arg EnrollCustomerParams) (EventsCustomerEnrollment, error) {
	row := q.db.QueryRowContext(ctx, enrollCustomer, arg.CustomerID, arg.EventID, arg.CheckedInAt)
	var i EventsCustomerEnrollment
	err := row.Scan(
		&i.ID,
		&i.CustomerID,
		&i.EventID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.CheckedInAt,
		&i.IsCancelled,
		&i.PaymentStatus,
		&i.PaymentExpiredAt,
	)
	return i, err
}

const getEventIsFull = `-- name: GetEventIsFull :one
SELECT COUNT(ce.customer_id) >= COALESCE(t.capacity, p.capacity)::boolean AS is_full
FROM events.events e
         JOIN program.programs p ON e.program_id = p.id
         LEFT JOIN athletic.teams t ON e.team_id = t.id
LEFT JOIN events.customer_enrollment ce ON e.id = ce.event_id
WHERE e.id = $1
GROUP BY e.id, t.capacity, p.capacity
`

func (q *Queries) GetEventIsFull(ctx context.Context, eventID uuid.UUID) (bool, error) {
	row := q.db.QueryRowContext(ctx, getEventIsFull, eventID)
	var is_full bool
	err := row.Scan(&is_full)
	return is_full, err
}

const unEnrollCustomer = `-- name: UnEnrollCustomer :execrows
UPDATE events.customer_enrollment
SET is_cancelled = true
WHERE customer_id = $1
  AND event_id = $2
`

type UnEnrollCustomerParams struct {
	CustomerID uuid.UUID `json:"customer_id"`
	EventID    uuid.UUID `json:"event_id"`
}

func (q *Queries) UnEnrollCustomer(ctx context.Context, arg UnEnrollCustomerParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, unEnrollCustomer, arg.CustomerID, arg.EventID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
