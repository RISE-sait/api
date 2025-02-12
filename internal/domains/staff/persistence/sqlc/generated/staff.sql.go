// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: staff.sql

package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const deleteStaff = `-- name: DeleteStaff :execrows
DELETE FROM staff WHERE id = $1
`

func (q *Queries) DeleteStaff(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteStaff, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getStaffByID = `-- name: GetStaffByID :one
SELECT s.id, is_active, created_at, updated_at, role_id, sr.id, role_name, sr.role_name FROM staff s 
JOIN staff_roles sr ON staff.role_id = staff_roles.id
WHERE s.id = $1
`

type GetStaffByIDRow struct {
	ID         uuid.UUID `json:"id"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	RoleID     uuid.UUID `json:"role_id"`
	ID_2       uuid.UUID `json:"id_2"`
	RoleName   string    `json:"role_name"`
	RoleName_2 string    `json:"role_name_2"`
}

func (q *Queries) GetStaffByID(ctx context.Context, id uuid.UUID) (GetStaffByIDRow, error) {
	row := q.db.QueryRowContext(ctx, getStaffByID, id)
	var i GetStaffByIDRow
	err := row.Scan(
		&i.ID,
		&i.IsActive,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.RoleID,
		&i.ID_2,
		&i.RoleName,
		&i.RoleName_2,
	)
	return i, err
}

const getStaffs = `-- name: GetStaffs :many
SELECT s.id, s.is_active, s.created_at, s.updated_at, s.role_id, sr.role_name FROM staff s
JOIN staff_roles sr ON s.role_id = sr.id
WHERE
(role_id = $1 OR $1 IS NULL)
`

type GetStaffsRow struct {
	ID        uuid.UUID `json:"id"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	RoleID    uuid.UUID `json:"role_id"`
	RoleName  string    `json:"role_name"`
}

func (q *Queries) GetStaffs(ctx context.Context, roleID uuid.NullUUID) ([]GetStaffsRow, error) {
	rows, err := q.db.QueryContext(ctx, getStaffs, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetStaffsRow
	for rows.Next() {
		var i GetStaffsRow
		if err := rows.Scan(
			&i.ID,
			&i.IsActive,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.RoleID,
			&i.RoleName,
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

const updateStaff = `-- name: UpdateStaff :one
WITH updated_staff AS (
    UPDATE staff s
    SET
        role_id = $1,
        is_active = $2
    WHERE s.id = $3
    RETURNING id, is_active, created_at, updated_at, role_id
)
SELECT us.id, us.is_active, us.created_at, us.updated_at, us.role_id, sr.role_name
FROM updated_staff us
JOIN staff_roles sr ON us.role_id = sr.id
`

type UpdateStaffParams struct {
	RoleID   uuid.UUID `json:"role_id"`
	IsActive bool      `json:"is_active"`
	ID       uuid.UUID `json:"id"`
}

type UpdateStaffRow struct {
	ID        uuid.UUID `json:"id"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	RoleID    uuid.UUID `json:"role_id"`
	RoleName  string    `json:"role_name"`
}

func (q *Queries) UpdateStaff(ctx context.Context, arg UpdateStaffParams) (UpdateStaffRow, error) {
	row := q.db.QueryRowContext(ctx, updateStaff, arg.RoleID, arg.IsActive, arg.ID)
	var i UpdateStaffRow
	err := row.Scan(
		&i.ID,
		&i.IsActive,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.RoleID,
		&i.RoleName,
	)
	return i, err
}
