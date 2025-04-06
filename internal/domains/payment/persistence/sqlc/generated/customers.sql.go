// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: customers.sql

package db_payment

import (
	"context"

	"github.com/google/uuid"
)

const getCustomersTeam = `-- name: GetCustomersTeam :one
SELECT t.id
FROM athletic.athletes a
         LEFT JOIN athletic.teams t ON a.team_id = t.id
WHERE a.id = $1
`

func (q *Queries) GetCustomersTeam(ctx context.Context, customerID uuid.UUID) (uuid.NullUUID, error) {
	row := q.db.QueryRowContext(ctx, getCustomersTeam, customerID)
	var id uuid.NullUUID
	err := row.Scan(&id)
	return id, err
}

const isCustomerExist = `-- name: IsCustomerExist :one
SELECT EXISTS(SELECT 1 FROM users.users WHERE id = $1)
`

func (q *Queries) IsCustomerExist(ctx context.Context, id uuid.UUID) (bool, error) {
	row := q.db.QueryRowContext(ctx, isCustomerExist, id)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}
