-- name: UpsertPushToken :one
INSERT INTO notifications.push_tokens (user_id, expo_push_token, device_type)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, device_type)
DO UPDATE SET 
    expo_push_token = EXCLUDED.expo_push_token,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetPushTokensByUserID :many
SELECT * FROM notifications.push_tokens 
WHERE user_id = $1;

-- name: GetPushTokensByTeamID :many
SELECT DISTINCT pt.* FROM notifications.push_tokens pt
JOIN users.users u ON pt.user_id = u.id
WHERE u.id IN (
    -- Team athletes (id column references users.users.id)
    SELECT a.id FROM athletic.athletes a WHERE a.team_id = $1
    UNION
    -- Team coach
    SELECT t.coach_id FROM athletic.teams t WHERE t.id = $1
);

-- name: DeletePushToken :exec
DELETE FROM notifications.push_tokens 
WHERE user_id = $1 AND device_type = $2;