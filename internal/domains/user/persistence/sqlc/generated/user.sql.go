// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: user.sql

package db_user

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const createAthleteInfo = `-- name: CreateAthleteInfo :execrows
INSERT INTO athletic.athletes (id, rebounds, assists, losses, wins, points)
VALUES ($1, $2, $3, $4, $5, $6)
`

type CreateAthleteInfoParams struct {
	ID       uuid.UUID `json:"id"`
	Rebounds int32     `json:"rebounds"`
	Assists  int32     `json:"assists"`
	Losses   int32     `json:"losses"`
	Wins     int32     `json:"wins"`
	Points   int32     `json:"points"`
}

func (q *Queries) CreateAthleteInfo(ctx context.Context, arg CreateAthleteInfoParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, createAthleteInfo,
		arg.ID,
		arg.Rebounds,
		arg.Assists,
		arg.Losses,
		arg.Wins,
		arg.Points,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

const getAthleteInfoByUserID = `-- name: GetAthleteInfoByUserID :one
SELECT id, wins, losses, points, steals, assists, rebounds, created_at, updated_at, team_id
FROM athletic.athletes
WHERE id = $1
limit 1
`

func (q *Queries) GetAthleteInfoByUserID(ctx context.Context, id uuid.UUID) (AthleticAthlete, error) {
	row := q.db.QueryRowContext(ctx, getAthleteInfoByUserID, id)
	var i AthleticAthlete
	err := row.Scan(
		&i.ID,
		&i.Wins,
		&i.Losses,
		&i.Points,
		&i.Steals,
		&i.Assists,
		&i.Rebounds,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.TeamID,
	)
	return i, err
}

const getChildren = `-- name: GetChildren :many
SELECT children.id, children.hubspot_id, children.country_alpha2_code, children.gender, children.first_name, children.last_name, children.age, children.parent_id, children.phone, children.email, children.has_marketing_email_consent, children.has_sms_consent, children.created_at, children.updated_at
FROM users.users parents
         JOIN users.users children
              ON parents.id = children.parent_id
WHERE parents.id = $1
`

func (q *Queries) GetChildren(ctx context.Context, id uuid.UUID) ([]UsersUser, error) {
	rows, err := q.db.QueryContext(ctx, getChildren, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []UsersUser
	for rows.Next() {
		var i UsersUser
		if err := rows.Scan(
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

const getCustomers = `-- name: GetCustomers :many
SELECT
    u.id, u.hubspot_id, u.country_alpha2_code, u.gender, u.first_name, u.last_name, u.age, u.parent_id, u.phone, u.email, u.has_marketing_email_consent, u.has_sms_consent, u.created_at, u.updated_at,
    -- Include other user fields you need
    m.name AS membership_name,
    mp.id AS membership_plan_id,
    mp.name AS membership_plan_name,
    cmp.start_date AS membership_start_date,
    cmp.renewal_date AS membership_plan_renewal_date,
    a.points,
    a.wins,
    a.losses,
    a.assists,
    a.rebounds,
    a.steals
FROM users.users u
         LEFT JOIN public.customer_membership_plans cmp ON (
    cmp.customer_id = u.id AND
    cmp.start_date = (
        SELECT MAX(start_date)
        FROM public.customer_membership_plans
        WHERE customer_id = u.id
    )
    )
         LEFT JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
         LEFT JOIN membership.memberships m ON m.id = mp.membership_id
         LEFT JOIN athletic.athletes a ON u.id = a.id
LIMIT $2 OFFSET $1
`

type GetCustomersParams struct {
	Offset int32 `json:"offset"`
	Limit  int32 `json:"limit"`
}

type GetCustomersRow struct {
	ID                        uuid.UUID      `json:"id"`
	HubspotID                 sql.NullString `json:"hubspot_id"`
	CountryAlpha2Code         string         `json:"country_alpha2_code"`
	Gender                    sql.NullString `json:"gender"`
	FirstName                 string         `json:"first_name"`
	LastName                  string         `json:"last_name"`
	Age                       int32          `json:"age"`
	ParentID                  uuid.NullUUID  `json:"parent_id"`
	Phone                     sql.NullString `json:"phone"`
	Email                     sql.NullString `json:"email"`
	HasMarketingEmailConsent  bool           `json:"has_marketing_email_consent"`
	HasSmsConsent             bool           `json:"has_sms_consent"`
	CreatedAt                 time.Time      `json:"created_at"`
	UpdatedAt                 time.Time      `json:"updated_at"`
	MembershipName            sql.NullString `json:"membership_name"`
	MembershipPlanID          uuid.NullUUID  `json:"membership_plan_id"`
	MembershipPlanName        sql.NullString `json:"membership_plan_name"`
	MembershipStartDate       sql.NullTime   `json:"membership_start_date"`
	MembershipPlanRenewalDate sql.NullTime   `json:"membership_plan_renewal_date"`
	Points                    sql.NullInt32  `json:"points"`
	Wins                      sql.NullInt32  `json:"wins"`
	Losses                    sql.NullInt32  `json:"losses"`
	Assists                   sql.NullInt32  `json:"assists"`
	Rebounds                  sql.NullInt32  `json:"rebounds"`
	Steals                    sql.NullInt32  `json:"steals"`
}

func (q *Queries) GetCustomers(ctx context.Context, arg GetCustomersParams) ([]GetCustomersRow, error) {
	rows, err := q.db.QueryContext(ctx, getCustomers, arg.Offset, arg.Limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetCustomersRow
	for rows.Next() {
		var i GetCustomersRow
		if err := rows.Scan(
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
			&i.MembershipName,
			&i.MembershipPlanID,
			&i.MembershipPlanName,
			&i.MembershipStartDate,
			&i.MembershipPlanRenewalDate,
			&i.Points,
			&i.Wins,
			&i.Losses,
			&i.Assists,
			&i.Rebounds,
			&i.Steals,
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

const getMembershipPlansByCustomer = `-- name: GetMembershipPlansByCustomer :many
SELECT cmp.id, cmp.customer_id, cmp.membership_plan_id, cmp.start_date, cmp.renewal_date, cmp.status, cmp.created_at, cmp.updated_at, m.name as membership_name
FROM public.customer_membership_plans cmp
         JOIN membership.membership_plans mp ON cmp.membership_plan_id = mp.id
         JOIN membership.memberships m ON m.id = mp.membership_id
WHERE cmp.customer_id = $1
`

type GetMembershipPlansByCustomerRow struct {
	ID               uuid.UUID        `json:"id"`
	CustomerID       uuid.UUID        `json:"customer_id"`
	MembershipPlanID uuid.UUID        `json:"membership_plan_id"`
	StartDate        time.Time        `json:"start_date"`
	RenewalDate      sql.NullTime     `json:"renewal_date"`
	Status           MembershipStatus `json:"status"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	MembershipName   string           `json:"membership_name"`
}

func (q *Queries) GetMembershipPlansByCustomer(ctx context.Context, customerID uuid.UUID) ([]GetMembershipPlansByCustomerRow, error) {
	rows, err := q.db.QueryContext(ctx, getMembershipPlansByCustomer, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetMembershipPlansByCustomerRow
	for rows.Next() {
		var i GetMembershipPlansByCustomerRow
		if err := rows.Scan(
			&i.ID,
			&i.CustomerID,
			&i.MembershipPlanID,
			&i.StartDate,
			&i.RenewalDate,
			&i.Status,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.MembershipName,
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

const updateAthleteStats = `-- name: UpdateAthleteStats :execrows
UPDATE athletic.athletes
SET wins       = COALESCE($1, wins),
    losses     = COALESCE($2, losses),
    points     = COALESCE($3, points),
    steals     = COALESCE($4, steals),
    assists    = COALESCE($5, assists),
    rebounds   = COALESCE($6, rebounds),
    updated_at = NOW()
WHERE id = $7
`

type UpdateAthleteStatsParams struct {
	Wins     sql.NullInt32 `json:"wins"`
	Losses   sql.NullInt32 `json:"losses"`
	Points   sql.NullInt32 `json:"points"`
	Steals   sql.NullInt32 `json:"steals"`
	Assists  sql.NullInt32 `json:"assists"`
	Rebounds sql.NullInt32 `json:"rebounds"`
	ID       uuid.UUID     `json:"id"`
}

func (q *Queries) UpdateAthleteStats(ctx context.Context, arg UpdateAthleteStatsParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, updateAthleteStats,
		arg.Wins,
		arg.Losses,
		arg.Points,
		arg.Steals,
		arg.Assists,
		arg.Rebounds,
		arg.ID,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
