// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: staff_queries.sql

package db_identity

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const createApprovedStaff = `-- name: CreateApprovedStaff :one
INSERT INTO staff.staff (id, role_id, is_active)
VALUES ($1,
        (SELECT id from staff.staff_roles where role_name = $2), $3)
RETURNING id, is_active, created_at, updated_at, role_id
`

type CreateApprovedStaffParams struct {
	ID       uuid.UUID `json:"id"`
	RoleName string    `json:"role_name"`
	IsActive bool      `json:"is_active"`
}

func (q *Queries) CreateApprovedStaff(ctx context.Context, arg CreateApprovedStaffParams) (StaffStaff, error) {
	row := q.db.QueryRowContext(ctx, createApprovedStaff, arg.ID, arg.RoleName, arg.IsActive)
	var i StaffStaff
	err := row.Scan(
		&i.ID,
		&i.IsActive,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.RoleID,
	)
	return i, err
}

const getStaffById = `-- name: GetStaffById :one
SELECT s.id, s.is_active, s.created_at, s.updated_at, s.role_id, sr.role_name, u.hubspot_id
FROM staff.staff s
         JOIN users.users u ON s.id = u.id
         JOIN staff.staff_roles sr ON s.role_id = sr.id
WHERE u.id = $1
`

type GetStaffByIdRow struct {
	ID        uuid.UUID      `json:"id"`
	IsActive  bool           `json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	RoleID    uuid.UUID      `json:"role_id"`
	RoleName  string         `json:"role_name"`
	HubspotID sql.NullString `json:"hubspot_id"`
}

func (q *Queries) GetStaffById(ctx context.Context, id uuid.UUID) (GetStaffByIdRow, error) {
	row := q.db.QueryRowContext(ctx, getStaffById, id)
	var i GetStaffByIdRow
	err := row.Scan(
		&i.ID,
		&i.IsActive,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.RoleID,
		&i.RoleName,
		&i.HubspotID,
	)
	return i, err
}

const getStaffRoles = `-- name: GetStaffRoles :many
SELECT id, role_name, created_at, updated_at
FROM staff.staff_roles
`

func (q *Queries) GetStaffRoles(ctx context.Context) ([]StaffStaffRole, error) {
	rows, err := q.db.QueryContext(ctx, getStaffRoles)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []StaffStaffRole
	for rows.Next() {
		var i StaffStaffRole
		if err := rows.Scan(
			&i.ID,
			&i.RoleName,
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
