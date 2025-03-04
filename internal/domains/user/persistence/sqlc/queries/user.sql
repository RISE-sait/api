-- name: GetUserIDByHubSpotId :one
SELECT id FROM users.users WHERE hubspot_id = $1;