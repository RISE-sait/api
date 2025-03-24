// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: memberships.sql

package db_seed

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
)

const insertClientsMembershipPlans = `-- name: InsertClientsMembershipPlans :exec
WITH prepared_data as (SELECT unnest($1::text[])       as customer_email,
                              unnest($2::text[])     as membership_plan_name,
                              unnest($3::timestamptz[])          as start_date,
                              unnest($4::timestamptz[]) as renewal_date)
INSERT INTO public.customer_membership_plans (customer_id, membership_plan_id, start_date, renewal_date)
SELECT u.id,
mp.id,
p.start_date,
       NULLIF(p.renewal_date, '1970-01-01 00:00:00+00'::timestamptz)
FROM prepared_data p
                 JOIN users.users u ON u.email = p.customer_email
JOIN membership.membership_plans mp ON mp.name = membership_plan_name
`

type InsertClientsMembershipPlansParams struct {
	CustomerEmailArray []string    `json:"customer_email_array"`
	MembershipPlanName []string    `json:"membership_plan_name"`
	StartDateArray     []time.Time `json:"start_date_array"`
	RenewalDateArray   []time.Time `json:"renewal_date_array"`
}

func (q *Queries) InsertClientsMembershipPlans(ctx context.Context, arg InsertClientsMembershipPlansParams) error {
	_, err := q.db.ExecContext(ctx, insertClientsMembershipPlans,
		pq.Array(arg.CustomerEmailArray),
		pq.Array(arg.MembershipPlanName),
		pq.Array(arg.StartDateArray),
		pq.Array(arg.RenewalDateArray),
	)
	return err
}

const insertCourseMembershipsEligibility = `-- name: InsertCourseMembershipsEligibility :exec
INSERT INTO public.program_membership (program_id, membership_id, is_eligible, price_per_booking)
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

const insertMembershipPlans = `-- name: InsertMembershipPlans :exec
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

func (q *Queries) InsertMembershipPlans(ctx context.Context, arg InsertMembershipPlansParams) error {
	_, err := q.db.ExecContext(ctx, insertMembershipPlans,
		pq.Array(arg.NameArray),
		pq.Array(arg.PriceArray),
		pq.Array(arg.JoiningFeeArray),
		pq.Array(arg.AutoRenewArray),
		pq.Array(arg.MembershipNameArray),
		pq.Array(arg.PaymentFrequencyArray),
		pq.Array(arg.AmtPeriodsArray),
	)
	return err
}

const insertMemberships = `-- name: InsertMemberships :exec
INSERT INTO membership.memberships (name, description)
VALUES (unnest($1::text[]), unnest($2::text[]))
RETURNING id
`

type InsertMembershipsParams struct {
	NameArray        []string `json:"name_array"`
	DescriptionArray []string `json:"description_array"`
}

func (q *Queries) InsertMemberships(ctx context.Context, arg InsertMembershipsParams) error {
	_, err := q.db.ExecContext(ctx, insertMemberships, pq.Array(arg.NameArray), pq.Array(arg.DescriptionArray))
	return err
}

const insertPracticeMembershipsEligibility = `-- name: InsertPracticeMembershipsEligibility :exec
WITH prepared_data as (SELECT unnest($1::text[])       as practice_names,
                              unnest($2::text[])     as membership_names,
                              unnest($3::bool[])          as is_eligible,
                              unnest($4::numeric[]) as price_per_booking)
INSERT
INTO public.program_membership (program_id, membership_id, is_eligible, price_per_booking)
SELECT p.id,
       m.id,
       is_eligible,
       CASE
           WHEN is_eligible = false THEN NULL::numeric
           ELSE price_per_booking
           END AS price_per_booking
FROM prepared_data
         JOIN membership.memberships m ON m.name = membership_names
         JOIN program.programs p ON p.name = practice_names
`

type InsertPracticeMembershipsEligibilityParams struct {
	PracticeNamesArray   []string          `json:"practice_names_array"`
	MembershipNamesArray []string          `json:"membership_names_array"`
	IsEligibleArray      []bool            `json:"is_eligible_array"`
	PricePerBookingArray []decimal.Decimal `json:"price_per_booking_array"`
}

func (q *Queries) InsertPracticeMembershipsEligibility(ctx context.Context, arg InsertPracticeMembershipsEligibilityParams) error {
	_, err := q.db.ExecContext(ctx, insertPracticeMembershipsEligibility,
		pq.Array(arg.PracticeNamesArray),
		pq.Array(arg.MembershipNamesArray),
		pq.Array(arg.IsEligibleArray),
		pq.Array(arg.PricePerBookingArray),
	)
	return err
}
