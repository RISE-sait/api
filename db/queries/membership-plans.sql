-- name: CreateMembershipPlan :execrows
INSERT INTO membership_plans (membership_id, name, price, payment_frequency, amt_periods)
VALUES ($1, $2, $3, $4, $5);

-- name: GetMembershipPlans :many
SELECT * 
FROM membership_plans
WHERE 
    ($1::UUID IS NULL OR $1::UUID = '00000000-0000-0000-0000-000000000000' OR membership_id = $1) AND
    ($2::UUID IS NULL OR $2::UUID = '00000000-0000-0000-0000-000000000000' OR id = $2);

-- name: UpdateMembershipPlan :execrows
UPDATE membership_plans
SET name = $1, price = $2, payment_frequency = $3, amt_periods = $4
WHERE membership_id = $5 AND id = $6;

-- name: DeleteMembershipPlan :execrows
DELETE FROM membership_plans WHERE membership_id = $1 AND id = $2;