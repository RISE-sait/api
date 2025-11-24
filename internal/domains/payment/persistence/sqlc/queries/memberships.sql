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
WHERE mp.stripe_price_id = $1 OR mp.stripe_joining_fee_id = $1;

-- name: GetMembershipPlanAmtPeriods :one
SELECT amt_periods
FROM membership.membership_plans
WHERE id = $1;

-- name: CheckCustomerActiveMembership :one
SELECT COUNT(*) as active_count
FROM users.customer_membership_plans
WHERE customer_id = $1 
  AND membership_plan_id = $2 
  AND status = 'active'
  AND (renewal_date IS NULL OR renewal_date > NOW());