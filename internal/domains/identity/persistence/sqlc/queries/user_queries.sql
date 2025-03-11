-- name: CreateUser :one
INSERT INTO users.users (hubspot_id, country_alpha2_code, email, age, phone, has_marketing_email_consent,
                         has_sms_consent, parent_id, first_name, last_name)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: UpdateUserHubspotId :execrows
UPDATE users.users
SET hubspot_id = $1
WHERE id = $2;

-- name: GetUserByUserID :one
SELECT *
FROM users.users
WHERE id = $1;

-- name: GetUserByHubSpotID :one
SELECT *
from users.users
WHERE hubspot_id = $1;