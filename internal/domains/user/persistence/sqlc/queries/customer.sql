-- name: UpdateCustomerStats :execrows
UPDATE users.users
SET
    wins = COALESCE(sqlc.narg('wins'), wins),
    losses = COALESCE(sqlc.narg('losses'), losses),
    points = COALESCE(sqlc.narg('points'), points),
    steals = COALESCE(sqlc.narg('steals'), steals),
    assists = COALESCE(sqlc.narg('assists'), assists),
    rebounds = COALESCE(sqlc.narg('rebounds'), rebounds),
    updated_at = NOW()
WHERE id = sqlc.arg('id');

-- name: GetCustomers :many
SELECT * FROM users.users
WHERE
    hubspot_id = ANY(sqlc.narg('hubspot_ids')::text[]) OR sqlc.narg('hubspot_ids') IS NULL;

-- name: GetMembershipPlansByCustomer :many
SELECT cmp.*, m.name as membership_name
FROM public.customer_membership_plans cmp
         JOIN membership.membership_plans mp ON cmp.membership_plan_id = mp.id
         JOIN membership.memberships m ON m.id = mp.membership_id
WHERE cmp.customer_id = $1;