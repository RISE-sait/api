-- name: GetUserIDByHubSpotId :one
SELECT id FROM users.users WHERE hubspot_id = $1;

-- name: InsertUser :execrows
INSERT INTO users.users (hubspot_id)
VALUES ($1);