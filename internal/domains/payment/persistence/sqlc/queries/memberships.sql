-- name: GetMembershipPlanJoiningRequirements :one
SELECT *
FROM membership.membership_plans
WHERE id = $1;

-- name: CreateCustomerMembershipPlan :exec
INSERT INTO users.customer_membership_plans (customer_id, membership_plan_id, status, start_date, renewal_date)
VALUES ($1, $2, $3, $4, $5);

-- name: GetMembershipPlanByStripePriceId :one
SELECT mp.id, mp.amt_periods
FROM membership.membership_plans mp
         LEFT JOIN membership.memberships m ON m.id = mp.membership_id
WHERE mp.stripe_price_id = $1;