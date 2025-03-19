// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: seed.sql

package db_seed

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

const insertAthletes = `-- name: InsertAthletes :many
INSERT
INTO users.athletes (id)
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

const insertClientsMembershipPlans = `-- name: InsertClientsMembershipPlans :many

INSERT INTO public.customer_membership_plans (customer_id, membership_plan_id, start_date, renewal_date)
VALUES (unnest($1::uuid[]),
        unnest($2::uuid[]),
        unnest($3::timestamptz[]),
        unnest($4::timestamptz[]))
RETURNING id
`

type InsertClientsMembershipPlansParams struct {
	CustomerID       []uuid.UUID `json:"customer_id"`
	PlansArray       []uuid.UUID `json:"plans_array"`
	StartDateArray   []time.Time `json:"start_date_array"`
	RenewalDateArray []time.Time `json:"renewal_date_array"`
}

// -- name: InsertClientsMembershipPlans :exec
// WITH prepared_data AS (SELECT unnest(@customer_id_array::uuid[])  AS customer_id,
//
//	unnest(
//	        ARRAY(
//	                SELECT CASE
//	                           WHEN membership_plan_id = '00000000-0000-0000-0000-000000000000'
//	                               THEN NULL
//	                           ELSE membership_plan_id
//	                           END
//	                FROM unnest(@membership_plan_id_array::uuid[]) AS membership_plan_id
//	        )
//	),
//	    unnest(
//	        ARRAY(
//	                SELECT CASE
//	                           WHEN start_date = '0001-01-01 00:00:00 UTC'
//	                               THEN NULL
//	                           ELSE start_date
//	                           END
//	                FROM unnest(@start_date_array::timestamptz[]) AS start_date
//	        )
//	)                                   AS start_date,
//	unnest(
//	        ARRAY(
//	                SELECT CASE
//	                           WHEN renewal_date = '0001-01-01 00:00:00 UTC'
//	                               THEN NULL
//	                           ELSE renewal_date
//	                           END
//	                FROM unnest(@renewal_date_array::timestamptz[]) AS renewal_date
//	        )
//	)                                   AS renewal_date)
//
// INSERT INTO public.customer_membership_plans (customer_id, membership_plan_id, start_date, renewal_date)
// VALUES (  customer_id, membership_plan_id, start_date, renewal_date);
func (q *Queries) InsertClientsMembershipPlans(ctx context.Context, arg InsertClientsMembershipPlansParams) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, insertClientsMembershipPlans,
		pq.Array(arg.CustomerID),
		pq.Array(arg.PlansArray),
		pq.Array(arg.StartDateArray),
		pq.Array(arg.RenewalDateArray),
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

const insertCourseMembershipsEligibility = `-- name: InsertCourseMembershipsEligibility :exec
INSERT INTO public.course_membership (course_id, membership_id, is_eligible, price_per_booking)
VALUES (unnest($1::uuid[]),
        unnest($2::uuid[]),
        unnest($3::bool[]),
        unnest($4::numeric[]))
`

type InsertCourseMembershipsEligibilityParams struct {
	CourseIDArray        []uuid.UUID       `json:"course_id_array"`
	MembershipIDArray    []uuid.UUID       `json:"membership_id_array"`
	IsEligibleArray      []bool            `json:"is_eligible_array"`
	PricePerBookingArray []decimal.Decimal `json:"price_per_booking_array"`
}

func (q *Queries) InsertCourseMembershipsEligibility(ctx context.Context, arg InsertCourseMembershipsEligibilityParams) error {
	_, err := q.db.ExecContext(ctx, insertCourseMembershipsEligibility,
		pq.Array(arg.CourseIDArray),
		pq.Array(arg.MembershipIDArray),
		pq.Array(arg.IsEligibleArray),
		pq.Array(arg.PricePerBookingArray),
	)
	return err
}

const insertCourses = `-- name: InsertCourses :many
INSERT INTO course.courses (name, description, capacity)
VALUES (unnest($1::text[]),
        unnest($2::text[]),
        unnest($3::int[]))
RETURNING id
`

type InsertCoursesParams struct {
	NameArray        []string `json:"name_array"`
	DescriptionArray []string `json:"description_array"`
	CapacityArray    []int32  `json:"capacity_array"`
}

func (q *Queries) InsertCourses(ctx context.Context, arg InsertCoursesParams) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, insertCourses, pq.Array(arg.NameArray), pq.Array(arg.DescriptionArray), pq.Array(arg.CapacityArray))
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

const insertCustomersEnrollments = `-- name: InsertCustomersEnrollments :many
WITH prepared_data AS (SELECT unnest($1::uuid[])  AS customer_id,
                              unnest($2::uuid[])     AS event_id,
                              unnest(
                                      ARRAY(
                                              SELECT CASE
                                                         WHEN checked_in_at = '0001-01-01 00:00:00 UTC'
                                                             THEN NULL
                                                         ELSE checked_in_at
                                                         END
                                              FROM unnest($3::timestamptz[]) AS checked_in_at
                                      )
                              )                                   AS checked_in_at,
                              unnest($4::bool[]) AS is_cancelled)
INSERT
INTO public.customer_enrollment(customer_id, event_id, checked_in_at, is_cancelled)
SELECT customer_id,
       event_id,
       checked_in_at,
       is_cancelled
FROM prepared_data
RETURNING id
`

type InsertCustomersEnrollmentsParams struct {
	CustomerIDArray  []uuid.UUID `json:"customer_id_array"`
	EventIDArray     []uuid.UUID `json:"event_id_array"`
	CheckedInAtArray []time.Time `json:"checked_in_at_array"`
	IsCancelledArray []bool      `json:"is_cancelled_array"`
}

func (q *Queries) InsertCustomersEnrollments(ctx context.Context, arg InsertCustomersEnrollmentsParams) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, insertCustomersEnrollments,
		pq.Array(arg.CustomerIDArray),
		pq.Array(arg.EventIDArray),
		pq.Array(arg.CheckedInAtArray),
		pq.Array(arg.IsCancelledArray),
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

const insertGames = `-- name: InsertGames :many
INSERT INTO public.games (name)
VALUES (unnest($1::text[]))
RETURNING id
`

func (q *Queries) InsertGames(ctx context.Context, nameArray []string) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, insertGames, pq.Array(nameArray))
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

const insertLocations = `-- name: InsertLocations :many
INSERT INTO location.locations (name, address)
VALUES (unnest($1::text[]), unnest($2::text[]))
RETURNING id
`

type InsertLocationsParams struct {
	NameArray    []string `json:"name_array"`
	AddressArray []string `json:"address_array"`
}

func (q *Queries) InsertLocations(ctx context.Context, arg InsertLocationsParams) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, insertLocations, pq.Array(arg.NameArray), pq.Array(arg.AddressArray))
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

const insertMembershipPlans = `-- name: InsertMembershipPlans :many
INSERT INTO membership.membership_plans (name, price, joining_fee, auto_renew, membership_id, payment_frequency,
                                         amt_periods)
SELECT name,
       price,
       joining_fee,
       auto_renew,
       (SELECT id FROM membership.memberships m WHERE m.name = membership_name),
       payment_frequency,
       amt_periods
FROM unnest($1::text[]) WITH ORDINALITY AS n(name, ord)
         JOIN
     unnest($2::numeric[]) WITH ORDINALITY AS p(price, ord) ON n.ord = p.ord
         JOIN
     unnest($3::numeric[]) WITH ORDINALITY AS j(joining_fee, ord) ON n.ord = j.ord
         JOIN
     unnest($4::boolean[]) WITH ORDINALITY AS a(auto_renew, ord) ON n.ord = a.ord
         JOIN
     unnest($5::text[]) WITH ORDINALITY AS m(membership_name, ord) ON n.ord = m.ord
         JOIN
     unnest($6::payment_frequency[]) WITH ORDINALITY AS f(payment_frequency, ord) ON n.ord = f.ord
         JOIN
     unnest($7::int[]) WITH ORDINALITY AS ap(amt_periods, ord) ON n.ord = ap.ord
RETURNING id
`

type InsertMembershipPlansParams struct {
	NameArray             []string           `json:"name_array"`
	PriceArray            []decimal.Decimal  `json:"price_array"`
	JoiningFeeArray       []decimal.Decimal  `json:"joining_fee_array"`
	AutoRenewArray        []bool             `json:"auto_renew_array"`
	MembershipNameArray   []string           `json:"membership_name_array"`
	PaymentFrequencyArray []PaymentFrequency `json:"payment_frequency_array"`
	AmtPeriodsArray       []int32            `json:"amt_periods_array"`
}

func (q *Queries) InsertMembershipPlans(ctx context.Context, arg InsertMembershipPlansParams) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, insertMembershipPlans,
		pq.Array(arg.NameArray),
		pq.Array(arg.PriceArray),
		pq.Array(arg.JoiningFeeArray),
		pq.Array(arg.AutoRenewArray),
		pq.Array(arg.MembershipNameArray),
		pq.Array(arg.PaymentFrequencyArray),
		pq.Array(arg.AmtPeriodsArray),
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

const insertMemberships = `-- name: InsertMemberships :many

INSERT INTO membership.memberships (name, description)
VALUES (unnest($1::text[]), unnest($2::text[]))
RETURNING id
`

type InsertMembershipsParams struct {
	NameArray        []string `json:"name_array"`
	DescriptionArray []string `json:"description_array"`
}

// -- name: InsertEvents :many
// INSERT INTO public.events (event_start_at, event_end_at, practice_id, course_id, game_id, location_id)
// SELECT unnest(@event_start_at_array::timestamptz[]),
//
//	unnest(@event_end_at_array::timestamptz[]),
//	unnest(
//	        ARRAY(
//	                SELECT CASE
//	                           WHEN practice_id = '00000000-0000-0000-0000-000000000000'
//	                               THEN NULL
//	                           ELSE practice_id
//	                           END
//	                FROM unnest(@practice_id_array::uuid[]) AS practice_id
//	        )
//	),
//	unnest(
//	        ARRAY(
//	                SELECT CASE
//	                           WHEN course_id = '00000000-0000-0000-0000-000000000000'
//	                               THEN NULL
//	                           ELSE course_id
//	                           END
//	                FROM unnest(@course_id_array::uuid[]) AS course_id
//	        )
//	),
//	unnest(
//	        ARRAY(
//	                SELECT CASE
//	                           WHEN game_id = '00000000-0000-0000-0000-000000000000'
//	                               THEN NULL
//	                           ELSE game_id
//	                           END
//	                FROM unnest(@game_id_array::uuid[]) AS game_id
//	        )
//	),
//	unnest(@location_id_array::uuid[])
//
// ON CONFLICT DO NOTHING
// RETURNING id;
func (q *Queries) InsertMemberships(ctx context.Context, arg InsertMembershipsParams) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, insertMemberships, pq.Array(arg.NameArray), pq.Array(arg.DescriptionArray))
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

const insertPracticeMembershipsEligibility = `-- name: InsertPracticeMembershipsEligibility :exec
INSERT INTO public.practice_membership (practice_id, membership_id, is_eligible, price_per_booking)
VALUES (unnest($1::uuid[]),
        unnest($2::uuid[]),
        unnest($3::bool[]),
        unnest($4::numeric[]))
`

type InsertPracticeMembershipsEligibilityParams struct {
	PracticeIDArray      []uuid.UUID       `json:"practice_id_array"`
	MembershipIDArray    []uuid.UUID       `json:"membership_id_array"`
	IsEligibleArray      []bool            `json:"is_eligible_array"`
	PricePerBookingArray []decimal.Decimal `json:"price_per_booking_array"`
}

func (q *Queries) InsertPracticeMembershipsEligibility(ctx context.Context, arg InsertPracticeMembershipsEligibilityParams) error {
	_, err := q.db.ExecContext(ctx, insertPracticeMembershipsEligibility,
		pq.Array(arg.PracticeIDArray),
		pq.Array(arg.MembershipIDArray),
		pq.Array(arg.IsEligibleArray),
		pq.Array(arg.PricePerBookingArray),
	)
	return err
}

const insertPractices = `-- name: InsertPractices :many
INSERT INTO public.practices (name, description, level, capacity)
VALUES (unnest($1::text[]),
        unnest($2::text[]),
        unnest($3::practice_level[]),
        unnest($4::int[]))
RETURNING id
`

type InsertPracticesParams struct {
	NameArray        []string        `json:"name_array"`
	DescriptionArray []string        `json:"description_array"`
	LevelArray       []PracticeLevel `json:"level_array"`
	CapacityArray    []int32         `json:"capacity_array"`
}

func (q *Queries) InsertPractices(ctx context.Context, arg InsertPracticesParams) ([]uuid.UUID, error) {
	rows, err := q.db.QueryContext(ctx, insertPractices,
		pq.Array(arg.NameArray),
		pq.Array(arg.DescriptionArray),
		pq.Array(arg.LevelArray),
		pq.Array(arg.CapacityArray),
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

const insertStaff = `-- name: InsertStaff :exec
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
INTO users.staff (id, is_active, role_id)
SELECT u.id,
       sd.is_active,
       sr.id
FROM staff_data sd
         JOIN
     users.users u ON u.email = sd.email
         JOIN
     users.staff_roles sr ON sr.role_name = sd.role_name
`

type InsertStaffParams struct {
	Emails        []string `json:"emails"`
	IsActiveArray []bool   `json:"is_active_array"`
	RoleNameArray []string `json:"role_name_array"`
}

func (q *Queries) InsertStaff(ctx context.Context, arg InsertStaffParams) error {
	_, err := q.db.ExecContext(ctx, insertStaff, pq.Array(arg.Emails), pq.Array(arg.IsActiveArray), pq.Array(arg.RoleNameArray))
	return err
}

const insertStaffRoles = `-- name: InsertStaffRoles :exec
INSERT INTO users.staff_roles (role_name)
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
                              unnest($4::int[])                             AS age,
                              unnest(
                                      ARRAY(
                                              SELECT CASE
                                                         WHEN parent_id = '00000000-0000-0000-0000-000000000000'
                                                             THEN NULL
                                                         ELSE parent_id
                                                         END
                                              FROM unnest($5::uuid[]) AS parent_id
                                      )
                              )                                                     AS parent_id,
                              unnest(
                                      ARRAY(
                                              SELECT CASE
                                                         WHEN gender = 'N'
                                                             THEN NULL
                                                         ELSE gender
                                                         END
                                              FROM unnest($6::char[]) AS gender
                                      )
                              ) AS gender,
                              unnest($7::text[])                          AS phone,
                              unnest($8::text[])                          AS email,
                              unnest($9::boolean[]) AS has_marketing_email_consent,
                              unnest($10::boolean[])             AS has_sms_consent)
INSERT
INTO users.users (country_alpha2_code,
                  first_name,
                  last_name,
                  age,
                  gender,
                  parent_id,
                  phone,
                  email,
                  has_marketing_email_consent,
                  has_sms_consent)
SELECT country_alpha2_code,
       first_name,
       last_name,
       age,
       gender,
       parent_id,
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
	AgeArray                      []int32     `json:"age_array"`
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
		pq.Array(arg.AgeArray),
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
