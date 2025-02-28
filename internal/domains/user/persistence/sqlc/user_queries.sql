-- name: GetUserIDByHubSpotId :one
SELECT id FROM users.users WHERE hubspot_id = $1;

-- name: GetUsers :many
SELECT * FROM users.users;

-- name: UpdateUserStats :execrows
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
