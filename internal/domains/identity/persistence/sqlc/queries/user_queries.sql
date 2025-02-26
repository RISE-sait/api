-- name: CreateUser :one
INSERT INTO users (hubspot_id) VALUES ($1)
RETURNING *;

-- name: UpdateUserHubspotId :execrows
UPDATE users
SET hubspot_id = $1
WHERE id = $2;

-- name: GetUserByHubSpotId :one
SELECT * FROM users WHERE hubspot_id = $1 LIMIT 1;