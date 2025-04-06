// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: memberships.sql

package db

import (
	"context"

	"github.com/google/uuid"
)

const createMembership = `-- name: CreateMembership :one
INSERT INTO membership.memberships (name, description, benefits)
VALUES ($1, $2, $3)
RETURNING id, name, description, benefits, created_at, updated_at
`

type CreateMembershipParams struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Benefits    string `json:"benefits"`
}

func (q *Queries) CreateMembership(ctx context.Context, arg CreateMembershipParams) (MembershipMembership, error) {
	row := q.db.QueryRowContext(ctx, createMembership, arg.Name, arg.Description, arg.Benefits)
	var i MembershipMembership
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Benefits,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteMembership = `-- name: DeleteMembership :execrows
DELETE FROM membership.memberships WHERE id = $1
`

func (q *Queries) DeleteMembership(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteMembership, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getMembershipById = `-- name: GetMembershipById :one
SELECT id, name, description, benefits, created_at, updated_at FROM membership.memberships WHERE id = $1
`

func (q *Queries) GetMembershipById(ctx context.Context, id uuid.UUID) (MembershipMembership, error) {
	row := q.db.QueryRowContext(ctx, getMembershipById, id)
	var i MembershipMembership
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Benefits,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const getMemberships = `-- name: GetMemberships :many
SELECT id, name, description, benefits, created_at, updated_at FROM membership.memberships
`

func (q *Queries) GetMemberships(ctx context.Context) ([]MembershipMembership, error) {
	rows, err := q.db.QueryContext(ctx, getMemberships)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []MembershipMembership
	for rows.Next() {
		var i MembershipMembership
		if err := rows.Scan(
			&i.ID,
			&i.Name,
			&i.Description,
			&i.Benefits,
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

const updateMembership = `-- name: UpdateMembership :one
UPDATE membership.memberships
SET name        = $1,
    description = $2,
    benefits    = $3,
    updated_at  = CURRENT_TIMESTAMP
WHERE id = $4
RETURNING id, name, description, benefits, created_at, updated_at
`

type UpdateMembershipParams struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Benefits    string    `json:"benefits"`
	ID          uuid.UUID `json:"id"`
}

func (q *Queries) UpdateMembership(ctx context.Context, arg UpdateMembershipParams) (MembershipMembership, error) {
	row := q.db.QueryRowContext(ctx, updateMembership,
		arg.Name,
		arg.Description,
		arg.Benefits,
		arg.ID,
	)
	var i MembershipMembership
	err := row.Scan(
		&i.ID,
		&i.Name,
		&i.Description,
		&i.Benefits,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}
