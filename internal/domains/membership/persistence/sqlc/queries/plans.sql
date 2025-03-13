-- name: CreateMembershipPlan :execrows
INSERT INTO membership.membership_plans (membership_id, name, price, payment_frequency,
                                         amt_periods, auto_renew, joining_fee)
VALUES ($1, $2, $3, $4,
        $5, $6, $7);

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
SET name              = $1,
    price             = $2,
    payment_frequency = $3,
    amt_periods       = $4,
    membership_id     = $5,
    auto_renew        = $6,
    joining_fee       = $7,
    updated_at        = CURRENT_TIMESTAMP
WHERE id = $8;

-- name: DeleteMembershipPlan :execrows
DELETE FROM membership.membership_plans WHERE id = $1;