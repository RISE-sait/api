-- name: UpdateAthleteStats :execrows
UPDATE users.athletes
SET wins       = COALESCE(sqlc.narg('wins'), wins),
    losses     = COALESCE(sqlc.narg('losses'), losses),
    points     = COALESCE(sqlc.narg('points'), points),
    steals     = COALESCE(sqlc.narg('steals'), steals),
    assists    = COALESCE(sqlc.narg('assists'), assists),
    rebounds   = COALESCE(sqlc.narg('rebounds'), rebounds),
    updated_at = NOW()
WHERE id = sqlc.arg('id');

-- name: GetCustomers :many
WITH latest_membership AS (SELECT cmp.customer_id,
                                  m.name         AS membership_name,
                                  cmp.start_date AS membership_start_date
                           FROM public.customer_membership_plans cmp
                                    JOIN
                                membership.membership_plans mp
                                ON mp.id = cmp.membership_plan_id
                                    JOIN
                                membership.memberships m
                                ON m.id = mp.membership_id
                           WHERE cmp.start_date = (SELECT MAX(cmp2.start_date)
                                                   FROM public.customer_membership_plans cmp2
                                                   WHERE cmp2.customer_id = cmp.customer_id))
SELECT u.*,
       lm.membership_name,      -- This will be NULL if no membership exists
       lm.membership_start_date, -- This will be NULL if no membership exists
       a.points,
       a.wins,
       a.losses,
       a.assists,
       a.rebounds,
       a.steals
FROM users.users u
         LEFT JOIN
     latest_membership lm
     ON lm.customer_id = u.id
         LEFT JOIN users.athletes a ON u.id = a.id
LIMIT $1 OFFSET $2;

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
INSERT INTO users.athletes (id, rebounds, assists, losses, wins, points)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetAthleteInfoByUserID :one
SELECT *
FROM users.athletes
WHERE id = $1
limit 1;
