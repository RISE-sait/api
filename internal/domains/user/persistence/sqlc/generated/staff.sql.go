// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: staff.sql

package db_user

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const createStaffRole = `-- name: CreateStaffRole :one
INSERT INTO staff.staff_roles (role_name)
VALUES ($1)
RETURNING id, role_name, created_at, updated_at
`

func (q *Queries) CreateStaffRole(ctx context.Context, roleName string) (StaffStaffRole, error) {
	row := q.db.QueryRowContext(ctx, createStaffRole, roleName)
	var i StaffStaffRole
	err := row.Scan(
		&i.ID,
		&i.RoleName,
		&i.CreatedAt,
		&i.UpdatedAt,
	)
	return i, err
}

const deleteStaff = `-- name: DeleteStaff :execrows
DELETE
FROM staff.staff
WHERE id = $1
`

func (q *Queries) DeleteStaff(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := q.db.ExecContext(ctx, deleteStaff, id)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getAvailableStaffRoles = `-- name: GetAvailableStaffRoles :many
SELECT id, role_name, created_at, updated_at
FROM staff.staff_roles
`

func (q *Queries) GetAvailableStaffRoles(ctx context.Context) ([]StaffStaffRole, error) {
	rows, err := q.db.QueryContext(ctx, getAvailableStaffRoles)
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

const getStaffs = `-- name: GetStaffs :many
SELECT s.is_active, u.id, u.hubspot_id, u.country_alpha2_code, u.gender, u.first_name, u.last_name, u.age, u.parent_id, u.phone, u.email, u.has_marketing_email_consent, u.has_sms_consent, u.created_at, u.updated_at, sr.role_name, cs.wins, cs.losses
FROM staff.staff s
JOIN users.users u ON u.id = s.id
JOIN staff.staff_roles sr ON s.role_id = sr.id
LEFT JOIN athletic.coach_stats cs ON s.id = cs.coach_id
WHERE (sr.role_name = $1 OR $1 IS NULL)
`

type GetStaffsRow struct {
	IsActive                 bool           `json:"is_active"`
	ID                       uuid.UUID      `json:"id"`
	HubspotID                sql.NullString `json:"hubspot_id"`
	CountryAlpha2Code        string         `json:"country_alpha2_code"`
	Gender                   sql.NullString `json:"gender"`
	FirstName                string         `json:"first_name"`
	LastName                 string         `json:"last_name"`
	Age                      int32          `json:"age"`
	ParentID                 uuid.NullUUID  `json:"parent_id"`
	Phone                    sql.NullString `json:"phone"`
	Email                    sql.NullString `json:"email"`
	HasMarketingEmailConsent bool           `json:"has_marketing_email_consent"`
	HasSmsConsent            bool           `json:"has_sms_consent"`
	CreatedAt                time.Time      `json:"created_at"`
	UpdatedAt                time.Time      `json:"updated_at"`
	RoleName                 string         `json:"role_name"`
	Wins                     sql.NullInt32  `json:"wins"`
	Losses                   sql.NullInt32  `json:"losses"`
}

func (q *Queries) GetStaffs(ctx context.Context, roleName sql.NullString) ([]GetStaffsRow, error) {
	rows, err := q.db.QueryContext(ctx, getStaffs, roleName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetStaffsRow
	for rows.Next() {
		var i GetStaffsRow
		if err := rows.Scan(
			&i.IsActive,
			&i.ID,
			&i.HubspotID,
			&i.CountryAlpha2Code,
			&i.Gender,
			&i.FirstName,
			&i.LastName,
			&i.Age,
			&i.ParentID,
			&i.Phone,
			&i.Email,
			&i.HasMarketingEmailConsent,
			&i.HasSmsConsent,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.RoleName,
			&i.Wins,
			&i.Losses,
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

const updateCoachStats = `-- name: UpdateCoachStats :execrows
UPDATE athletic.coach_stats
SET wins       = COALESCE($1, wins),
    losses     = COALESCE($2, losses),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $3
`

type UpdateCoachStatsParams struct {
	Wins   sql.NullInt32 `json:"wins"`
	Losses sql.NullInt32 `json:"losses"`
	ID     uuid.UUID     `json:"id"`
}

func (q *Queries) UpdateCoachStats(ctx context.Context, arg UpdateCoachStatsParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, updateCoachStats, arg.Wins, arg.Losses, arg.ID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const updateStaff = `-- name: UpdateStaff :execrows
UPDATE staff.staff s
    SET role_id = (SELECT id from staff.staff_roles sr WHERE sr.role_name = $1),
        is_active  = $2,
        updated_at = CURRENT_TIMESTAMP
WHERE s.id = $3
`

type UpdateStaffParams struct {
	RoleName string    `json:"role_name"`
	IsActive bool      `json:"is_active"`
	ID       uuid.UUID `json:"id"`
}

func (q *Queries) UpdateStaff(ctx context.Context, arg UpdateStaffParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, updateStaff, arg.RoleName, arg.IsActive, arg.ID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
