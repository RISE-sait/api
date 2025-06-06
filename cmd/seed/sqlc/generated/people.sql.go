// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: people.sql

package db_seed

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

const insertAthletes = `-- name: InsertAthletes :many
INSERT
INTO athletic.athletes (id)
VALUES (unnest($1::uuid[]))
RETURNING id
`

func (q *Queries) InsertAthletes(ctx context.Context, idArray []uuid.UUID) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, insertAthletes, pq.Array(idArray))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertStaff = `-- name: InsertStaff :many
WITH staff_data AS (SELECT e.email,
                           ia.is_active,
                           rn.role_name
                    FROM unnest($1::text[]) WITH ORDINALITY AS e(email, idx)
                             JOIN
                         unnest($2::bool[]) WITH ORDINALITY AS ia(is_active, idx)
                         ON e.idx = ia.idx
                             JOIN
                         unnest($3::text[]) WITH ORDINALITY AS rn(role_name, idx)
                         ON e.idx = rn.idx)
INSERT
INTO staff.staff (id, is_active, role_id)
SELECT u.id,
       sd.is_active,
       sr.id
FROM staff_data sd
         JOIN
     users.users u ON u.email = sd.email
         JOIN
     staff.staff_roles sr ON sr.role_name = sd.role_name
RETURNING id
`

type InsertStaffParams struct {
	Emails        []string `json:"emails"`
	IsActiveArray []bool   `json:"is_active_array"`
	RoleNameArray []string `json:"role_name_array"`
}

func (q *Queries) InsertStaff(ctx context.Context, arg InsertStaffParams) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, insertStaff, pq.Array(arg.Emails), pq.Array(arg.IsActiveArray), pq.Array(arg.RoleNameArray))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const insertStaffRoles = `-- name: InsertStaffRoles :exec
INSERT INTO staff.staff_roles (role_name)
VALUES ('admin'),
       ('superadmin'),
       ('coach'),
       ('instructor'),
       ('receptionist'),
       ('barber')
`

func (q *Queries) InsertStaffRoles(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, insertStaffRoles)
	return err
}

const insertUsers = `-- name: InsertUsers :many
WITH prepared_data AS (SELECT unnest($1::text[])            AS country_alpha2_code,
                              unnest($2::text[])                     AS first_name,
                              unnest($3::text[])                      AS last_name,
                              unnest($4::timestamptz[]) AS dob,
                              unnest($5::uuid[]) AS parent_id,
                              unnest($6::char[])    AS gender,
                              unnest($7::text[])                          AS phone,
                              unnest($8::text[])                          AS email,
                              unnest($9::boolean[]) AS has_marketing_email_consent,
                              unnest($10::boolean[])             AS has_sms_consent)
INSERT
INTO users.users (country_alpha2_code,
                  first_name,
                  last_name,
                  dob,
                  gender,
                  parent_id,
                  phone,
                  email,
                  has_marketing_email_consent,
                  has_sms_consent)
SELECT country_alpha2_code,
       first_name,
       last_name,
       dob,
       NULLIF(gender, 'N')                                       AS gender,    -- Replace 'N' with NULL
       NULLIF(parent_id, '00000000-0000-0000-0000-000000000000') AS parent_id, -- Replace default UUID with NULL
       phone,
       email,
       has_marketing_email_consent,
       has_sms_consent
FROM prepared_data
ON CONFLICT DO NOTHING
RETURNING id
`

type InsertUsersParams struct {
	CountryAlpha2CodeArray        []string    `json:"country_alpha2_code_array"`
	FirstNameArray                []string    `json:"first_name_array"`
	LastNameArray                 []string    `json:"last_name_array"`
	DobArray                      []time.Time `json:"dob_array"`
	ParentIDArray                 []uuid.UUID `json:"parent_id_array"`
	GenderArray                   []string    `json:"gender_array"`
	PhoneArray                    []string    `json:"phone_array"`
	EmailArray                    []string    `json:"email_array"`
	HasMarketingEmailConsentArray []bool      `json:"has_marketing_email_consent_array"`
	HasSmsConsentArray            []bool      `json:"has_sms_consent_array"`
}

func (q *Queries) InsertUsers(ctx context.Context, arg InsertUsersParams) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, insertUsers,
		pq.Array(arg.CountryAlpha2CodeArray),
		pq.Array(arg.FirstNameArray),
		pq.Array(arg.LastNameArray),
		pq.Array(arg.DobArray),
		pq.Array(arg.ParentIDArray),
		pq.Array(arg.GenderArray),
		pq.Array(arg.PhoneArray),
		pq.Array(arg.EmailArray),
		pq.Array(arg.HasMarketingEmailConsentArray),
		pq.Array(arg.HasSmsConsentArray),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		items = append(items, id)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateParents = `-- name: UpdateParents :execrows
UPDATE users.users
SET parent_id = (SELECT id from users.users WHERE email = 'parent@gmail.com')
WHERE email IN ('klintlee1@gmail.com', 'sukhdeepboparai2005@gmail.com')
`

func (q *Queries) UpdateParents(ctx context.Context) (int64, error) {
	result, err := q.db.ExecContext(ctx, updateParents)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
