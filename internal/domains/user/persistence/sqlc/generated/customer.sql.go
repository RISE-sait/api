// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: customer.sql

package db_user

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const addAthleteToTeam = `-- name: AddAthleteToTeam :execrows
UPDATE athletic.athletes
SET team_id = $1
WHERE id = $2
`

type AddAthleteToTeamParams struct {
	TeamID     uuid.NullUUID `json:"team_id"`
	CustomerID uuid.UUID     `json:"customer_id"`
}

func (q *Queries) AddAthleteToTeam(ctx context.Context, arg AddAthleteToTeamParams) (int64, error) {
	result, err := q.db.ExecContext(ctx, addAthleteToTeam, arg.TeamID, arg.CustomerID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

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

const getCustomer = `-- name: GetCustomer :one
SELECT u.id, u.hubspot_id, u.country_alpha2_code, u.gender, u.first_name, u.last_name, u.age, u.parent_id, u.phone, u.email, u.has_marketing_email_consent, u.has_sms_consent, u.created_at, u.updated_at,
       m.name           AS membership_name,
       mp.id            AS membership_plan_id,
       mp.name          AS membership_plan_name,
       cmp.start_date   AS membership_start_date,
       cmp.renewal_date AS membership_plan_renewal_date,
       a.points,
       a.wins,
       a.losses,
       a.assists,
       a.rebounds,
       a.steals
FROM users.users u
         LEFT JOIN users.customer_membership_plans cmp ON (
    cmp.customer_id = u.id AND
    cmp.start_date = (SELECT MAX(start_date)
                      FROM users.customer_membership_plans
                      WHERE customer_id = u.id)
    )
         LEFT JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
         LEFT JOIN membership.memberships m ON m.id = mp.membership_id
         LEFT JOIN athletic.athletes a ON u.id = a.id
WHERE (u.id = $1 OR $1 IS NULL)
  AND (u.email = $2 OR $2 IS NULL)
  AND NOT EXISTS (SELECT 1
                  FROM staff.staff s
                  WHERE s.id = u.id)
`

type GetCustomerParams struct {
	ID    uuid.NullUUID  `json:"id"`
	Email sql.NullString `json:"email"`
}

type GetCustomerRow struct {
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

func (q *Queries) GetCustomer(ctx context.Context, arg GetCustomerParams) (GetCustomerRow, error) {
	row := q.db.QueryRowContext(ctx, getCustomer, arg.ID, arg.Email)
	var i GetCustomerRow
	err := row.Scan(
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
	)
	return i, err
}

const getCustomers = `-- name: GetCustomers :many
SELECT u.id, u.hubspot_id, u.country_alpha2_code, u.gender, u.first_name, u.last_name, u.age, u.parent_id, u.phone, u.email, u.has_marketing_email_consent, u.has_sms_consent, u.created_at, u.updated_at,
       m.name           AS membership_name,
       mp.id            AS membership_plan_id,
       mp.name          AS membership_plan_name,
       cmp.start_date   AS membership_start_date,
       cmp.renewal_date AS membership_plan_renewal_date,
       a.points,
       a.wins,
       a.losses,
       a.assists,
       a.rebounds,
       a.steals
FROM users.users u
         LEFT JOIN users.customer_membership_plans cmp ON (
    cmp.customer_id = u.id AND
    cmp.start_date = (SELECT MAX(start_date)
                      FROM users.customer_membership_plans
                      WHERE customer_id = u.id)
    )
         LEFT JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
         LEFT JOIN membership.memberships m ON m.id = mp.membership_id
         LEFT JOIN athletic.athletes a ON u.id = a.id
WHERE (u.parent_id = $1 OR $1 IS NULL)
  AND NOT EXISTS (SELECT 1
                  FROM staff.staff s
                  WHERE s.id = u.id)
LIMIT $3 OFFSET $2
`

type GetCustomersParams struct {
	ParentID uuid.NullUUID `json:"parent_id"`
	Offset   int32         `json:"offset"`
	Limit    int32         `json:"limit"`
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
	rows, err := q.db.QueryContext(ctx, getCustomers, arg.ParentID, arg.Offset, arg.Limit)
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

const updateAthleteStats = `-- name: UpdateAthleteStats :execrows
UPDATE athletic.athletes
SET wins       = COALESCE($1, wins),
    losses     = COALESCE($2, losses),
    points     = COALESCE($3, points),
    steals     = COALESCE($4, steals),
    assists    = COALESCE($5, assists),
    rebounds   = COALESCE($6, rebounds),
    updated_at = current_timestamp
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
