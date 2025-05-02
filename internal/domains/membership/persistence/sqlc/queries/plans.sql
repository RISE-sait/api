-- name: CreateMembershipPlan :one
INSERT INTO membership.membership_plans (membership_id, name, stripe_joining_fee_id, stripe_price_id, amt_periods)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetMembershipPlanById :one
SELECT *
FROM membership.membership_plans
WHERE id = $1;

-- name: GetMembershipPlans :many
SELECT 
  mp.id,
  mp.membership_id,
  mp.name,
  mp.stripe_price_id,
  mp.stripe_joining_fee_id,
  mp.amt_periods,
  mp.unit_amount,
  mp.currency,
  mp.interval,
  mp.created_at,
  mp.updated_at
FROM membership.membership_plans mp
WHERE mp.membership_id = $1;


-- name: UpdateMembershipPlan :one
UPDATE membership.membership_plans
SET name              = $1,
    stripe_price_id       = $2,
    stripe_joining_fee_id = $3,
    amt_periods       = $4,
    membership_id     = $5,
    updated_at        = CURRENT_TIMESTAMP
WHERE id = $6
RETURNING *;

-- name: DeleteMembershipPlan :execrows
DELETE FROM membership.membership_plans WHERE id = $1;