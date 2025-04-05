-- name: UpdateAthleteStats :execrows
UPDATE athletic.athletes
SET wins       = COALESCE(sqlc.narg('wins'), wins),
    losses     = COALESCE(sqlc.narg('losses'), losses),
    points     = COALESCE(sqlc.narg('points'), points),
    steals     = COALESCE(sqlc.narg('steals'), steals),
    assists    = COALESCE(sqlc.narg('assists'), assists),
    rebounds   = COALESCE(sqlc.narg('rebounds'), rebounds),
    updated_at = current_timestamp
WHERE id = sqlc.arg('id');

-- name: GetCustomers :many
SELECT u.*,
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
         LEFT JOIN users.customer_membership_plans cmp ON (
    cmp.customer_id = u.id AND
    cmp.start_date = (SELECT MAX(start_date)
                      FROM users.customer_membership_plans
                      WHERE customer_id = u.id)
    )
         LEFT JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
         LEFT JOIN membership.memberships m ON m.id = mp.membership_id
         LEFT JOIN athletic.athletes a ON u.id = a.id
WHERE (u.parent_id = $1 OR $1 IS NULL)
  AND NOT EXISTS (SELECT 1
                  FROM staff.staff s
                  WHERE s.id = u.id)
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetCustomer :one
SELECT u.*,
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
         LEFT JOIN users.customer_membership_plans cmp ON (
    cmp.customer_id = u.id AND
    cmp.start_date = (SELECT MAX(start_date)
                      FROM users.customer_membership_plans
                      WHERE customer_id = u.id)
    )
         LEFT JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
         LEFT JOIN membership.memberships m ON m.id = mp.membership_id
         LEFT JOIN athletic.athletes a ON u.id = a.id
WHERE (u.id = sqlc.narg('id') OR sqlc.narg('id') IS NULL)
  AND (u.email = sqlc.narg('email') OR sqlc.narg('email') IS NULL)
  AND NOT EXISTS (SELECT 1
                  FROM staff.staff s
                  WHERE s.id = u.id);

-- name: CreateAthleteInfo :execrows
INSERT INTO athletic.athletes (id, rebounds, assists, losses, wins, points)
VALUES ($1, $2, $3, $4, $5, $6);