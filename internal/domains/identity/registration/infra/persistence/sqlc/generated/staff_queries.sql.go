// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: staff_queries.sql

package db

import (
	"context"
)

const createStaff = `-- name: CreateStaff :execrows
INSERT INTO staff (role, is_active) VALUES ($1, $2)
`

type CreateStaffParams struct {
	Role     StaffRoleEnum `json:"role"`
	IsActive bool          `json:"is_active"`
}

func (q *Queries) CreateStaff(ctx context.Context, arg CreateStaffParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, createStaff, arg.Role, arg.IsActive)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
