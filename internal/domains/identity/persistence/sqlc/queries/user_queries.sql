-- name: CreateUser :one
INSERT INTO users.users (hubspot_id, country_alpha2_code, email, age, phone, has_marketing_email_consent,
                         has_sms_consent, parent_id, first_name, last_name)
VALUES ($1, $2, $3, $4, $5,
        $6, $7, (SELECT pu.id from users.users pu WHERE sqlc.arg('parent_email') = pu.email), $8, $9)
RETURNING *;

-- name: CreateAthlete :execrows
INSERT INTO users.athletes (id)
VALUES ($1);

-- name: UpdateUserHubspotId :execrows
UPDATE users.users
SET hubspot_id = $1
WHERE id = $2;

-- name: GetUserByID :one
SELECT *
FROM users.users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT *
FROM users.users
WHERE email = $1;

-- name: GetIsUserAParent :one
SELECT COUNT(*) > 0
FROM users.users parents
         JOIN users.users children
              ON children.parent_id = parents.id
WHERE parents.id = $1;

-- name: GetIsActualParentChild :one
SELECT COUNT(*) > 0
FROM users.users parents
         JOIN users.users children
              ON children.parent_id = parents.id
WHERE parents.email = sqlc.arg('parent_email')
  AND children.id = sqlc.arg('child_id');

-- name: GetIsAthleteByID :one
SELECT COUNT(*) > 0
FROM users.athletes
WHERE id = $1;