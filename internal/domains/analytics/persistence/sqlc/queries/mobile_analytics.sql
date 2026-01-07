-- name: RecordMobileLogin :exec
UPDATE users.users
SET last_mobile_login_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: GetMobileUsageStats :one
SELECT
    COUNT(*) FILTER (WHERE last_mobile_login_at IS NOT NULL) AS total_mobile_users,
    COUNT(*) FILTER (WHERE last_mobile_login_at >= CURRENT_DATE) AS active_today,
    COUNT(*) FILTER (WHERE last_mobile_login_at >= CURRENT_DATE - INTERVAL '7 days') AS active_last_7_days,
    COUNT(*) FILTER (WHERE last_mobile_login_at >= CURRENT_DATE - INTERVAL '30 days') AS active_last_30_days,
    COUNT(*) AS total_users
FROM users.users
WHERE deleted_at IS NULL;

-- name: GetRecentMobileLogins :many
SELECT
    u.id,
    u.first_name,
    u.last_name,
    u.email,
    u.last_mobile_login_at
FROM users.users u
WHERE u.last_mobile_login_at IS NOT NULL
  AND u.deleted_at IS NULL
ORDER BY u.last_mobile_login_at DESC
LIMIT $1 OFFSET $2;

-- name: GetMobileLoginsByDateRange :many
SELECT
    DATE(last_mobile_login_at) AS login_date,
    COUNT(DISTINCT id) AS unique_users
FROM users.users
WHERE last_mobile_login_at >= $1
  AND last_mobile_login_at < $2
  AND deleted_at IS NULL
GROUP BY DATE(last_mobile_login_at)
ORDER BY login_date DESC;
