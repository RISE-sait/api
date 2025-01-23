-- name: CreateMembershipPlan :execrows
INSERT INTO membership_plans (membership_id, name, price, payment_frequency, amt_periods)
VALUES ($1, $2, $3, $4, $5);

-- name: GetMembershipPlansByMembershipId :many
SELECT * 
FROM membership_plans
WHERE 
    membership_id = $1;

-- name: UpdateMembershipPlan :execrows
UPDATE membership_plans
SET name = $1, price = $2, payment_frequency = $3, amt_periods = $4
WHERE membership_id = $5 AND id = $6;

-- name: DeleteMembershipPlan :execrows
DELETE FROM membership_plans WHERE membership_id = $1 AND id = $2;