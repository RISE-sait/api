// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: event_capacity_queries.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const getEventIsFull = `-- name: GetEventIsFull :one
SELECT
    COUNT(ce.customer_id) >= COALESCE(p.capacity, c.capacity) AS is_full
FROM events e
         LEFT JOIN customer_enrollment ce ON e.id = ce.event_id
         LEFT JOIN practices p ON e.practice_id = p.id
         LEFT JOIN courses c ON e.course_id = c.id
WHERE e.id = $1
GROUP BY e.id, e.practice_id, e.course_id, p.capacity, c.capacity
`

func (q *Queries) GetEventIsFull(ctx context.Context, eventID uuid.UUID) (bool, error) {
	row := q.db.QueryRowContext(ctx, getEventIsFull, eventID)
	var is_full bool
	err := row.Scan(&is_full)
	return is_full, err
}
