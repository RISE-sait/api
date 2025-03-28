-- name: UpdateAthleteStats :execrows
UPDATE athletic.athletes
SET wins       = COALESCE(sqlc.narg('wins'), wins),
    losses     = COALESCE(sqlc.narg('losses'), losses),
    points     = COALESCE(sqlc.narg('points'), points),
    steals     = COALESCE(sqlc.narg('steals'), steals),
    assists    = COALESCE(sqlc.narg('assists'), assists),
    rebounds   = COALESCE(sqlc.narg('rebounds'), rebounds),
    updated_at = NOW()
WHERE id = sqlc.arg('id');

-- name: GetCustomers :many
SELECT u.*,
       -- Include other user fields you need
       m.name           AS membership_name,
       mp.id            AS membership_plan_id,
       mp.name          AS membership_plan_name,
       cmp.start_date   AS membership_start_date,
       cmp.renewal_date AS membership_plan_renewal_date,
       a.points,
       a.wins,
       a.losses,
       a.assists,
       a.rebounds,
       a.steals
FROM users.users u
         LEFT JOIN public.customer_membership_plans cmp ON (
    cmp.customer_id = u.id AND
    cmp.start_date = (SELECT MAX(start_date)
                      FROM public.customer_membership_plans
                      WHERE customer_id = u.id)
    )
         LEFT JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
         LEFT JOIN membership.memberships m ON m.id = mp.membership_id
         LEFT JOIN athletic.athletes a ON u.id = a.id
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetChildren :many
SELECT children.*
FROM users.users parents
         JOIN users.users children
              ON parents.id = children.parent_id
WHERE parents.id = $1;

-- name: GetMembershipPlansByCustomer :many
SELECT cmp.*, m.name as membership_name
FROM public.customer_membership_plans cmp
         JOIN membership.membership_plans mp ON cmp.membership_plan_id = mp.id
         JOIN membership.memberships m ON m.id = mp.membership_id
WHERE cmp.customer_id = $1;

-- name: CreateAthleteInfo :execrows
INSERT INTO athletic.athletes (id, rebounds, assists, losses, wins, points)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetAthleteInfoByUserID :one
SELECT *
FROM athletic.athletes
WHERE id = $1
limit 1;
