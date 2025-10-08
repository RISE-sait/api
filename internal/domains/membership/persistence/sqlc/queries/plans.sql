-- name: CreateMembershipPlan :one
INSERT INTO membership.membership_plans (membership_id, name, stripe_joining_fee_id, stripe_price_id, amt_periods, joining_fee, credit_allocation, weekly_credit_limit)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
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
  mp.joining_fee,
  mp.credit_allocation,
  mp.weekly_credit_limit,
  mp.is_visible,
  mp.created_at,
  mp.updated_at
FROM membership.membership_plans mp
WHERE mp.membership_id = $1
  AND mp.is_visible = true;


-- name: UpdateMembershipPlan :one
UPDATE membership.membership_plans
SET name              = $1,
    stripe_price_id       = $2,
    stripe_joining_fee_id = $3,
    amt_periods       = $4,
    membership_id     = $5,
    joining_fee       = $6,
    credit_allocation = $7,
    weekly_credit_limit = $8,
    updated_at        = CURRENT_TIMESTAMP
WHERE id = $9
RETURNING *;

-- name: DeleteMembershipPlan :execrows
DELETE FROM membership.membership_plans WHERE id = $1;

-- name: ToggleMembershipPlanVisibility :one
UPDATE membership.membership_plans
SET is_visible = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: GetAllMembershipPlansAdmin :many
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
  mp.joining_fee,
  mp.credit_allocation,
  mp.weekly_credit_limit,
  mp.is_visible,
  mp.created_at,
  mp.updated_at
FROM membership.membership_plans mp
WHERE mp.membership_id = $1;