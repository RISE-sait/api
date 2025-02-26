-- name: GetUserIDByHubSpotId :one
SELECT id FROM users WHERE hubspot_id = $1 LIMIT 1;