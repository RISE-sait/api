-- name: CreateMembershipPlan :execrows
INSERT INTO membership.membership_plans (membership_id, name, price, payment_frequency, amt_periods)
VALUES ($1, $2, $3, $4, $5);

-- name: GetMembershipPlanById :one
SELECT *
FROM membership.membership_plans
WHERE id = $1;

-- name: GetMembershipPlans :many
SELECT * 
FROM membership.membership_plans mp
WHERE mp.membership_id = $1;

-- name: UpdateMembershipPlan :execrows
UPDATE membership.membership_plans
SET name = $1, price = $2, payment_frequency = $3, amt_periods = $4, membership_id = $5
WHERE id = $6;

-- name: DeleteMembershipPlan :execrows
DELETE FROM membership.membership_plans WHERE id = $1;