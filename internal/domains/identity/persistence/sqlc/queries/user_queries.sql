-- name: CreateUser :one
INSERT INTO users.users (hubspot_id, country_alpha2_code, email, dob, phone, has_marketing_email_consent,
                         has_sms_consent, parent_id, first_name, last_name)
VALUES ($1, $2, $3, $4, $5,
        $6, $7, (SELECT pu.id from users.users pu WHERE sqlc.arg('parent_email') = pu.email), $8, $9)
RETURNING *;

-- name: CreateAthlete :exec
INSERT INTO athletic.athletes (id)
VALUES ($1);

-- name: UpdateUserHubspotId :execrows
UPDATE users.users
SET hubspot_id = $1
WHERE id = $2;

-- name: GetUserByIdOrEmail :one
WITH u
         as (SELECT *
             FROM users.users u2
             WHERE (u2.id = sqlc.narg('id') OR sqlc.narg('id') IS NULL)
               AND (u2.email = sqlc.narg('email') OR sqlc.narg('email') IS NULL)
             LIMIT 1),
     latest_cmp AS (SELECT DISTINCT ON (customer_id) *
                    FROM users.customer_membership_plans
                    WHERE customer_id = (SELECT id FROM u)
                    ORDER BY customer_id, start_date DESC)
SELECT u.*,
       mp.name          as membership_plan_name,
       cmp.start_date   as membership_plan_start_date,
       cmp.renewal_date as membership_plan_renewal_date,
       m.name        as membership_name,
       m.description as membership_description,
       m.benefits    as membership_benefits,

       a.points,
       a.wins,
       a.losses,
       a.assists,
       a.rebounds,
       a.steals,
       a.photo_url as athlete_photo_url
from u
         LEFT JOIN
     latest_cmp cmp ON cmp.customer_id = u.id
         LEFT JOIN membership.membership_plans mp ON mp.id = cmp.membership_plan_id
         LEFT JOIN membership.memberships m ON m.id = mp.membership_id
         LEFT JOIN athletic.athletes a ON u.id = a.id;

-- name: GetIsUserAParent :one
SELECT COUNT(*) > 0
FROM users.users
WHERE parent_id = sqlc.arg('parent_id');

-- name: GetIsActualParentChild :one
SELECT COUNT(*) > 0
FROM users.users
WHERE id = sqlc.arg('child_id')
  AND parent_id = sqlc.arg('parent_id');