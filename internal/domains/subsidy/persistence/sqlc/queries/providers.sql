-- name: CreateProvider :one
INSERT INTO subsidies.providers (name, contact_email, contact_phone, is_active)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetProvider :one
SELECT * FROM subsidies.providers
WHERE id = $1;

-- name: GetProviderByName :one
SELECT * FROM subsidies.providers
WHERE name = $1;

-- name: ListProviders :many
SELECT * FROM subsidies.providers
WHERE (sqlc.narg('is_active')::boolean IS NULL OR is_active = sqlc.narg('is_active'))
ORDER BY name ASC;

-- name: UpdateProvider :one
UPDATE subsidies.providers
SET
    name = COALESCE($2, name),
    contact_email = COALESCE($3, contact_email),
    contact_phone = COALESCE($4, contact_phone),
    is_active = COALESCE($5, is_active),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeleteProvider :exec
DELETE FROM subsidies.providers
WHERE id = $1;

-- name: GetProviderStats :one
SELECT
    p.id,
    p.name,
    COUNT(DISTINCT cs.id) as total_subsidies,
    COALESCE(SUM(cs.approved_amount), 0) as total_amount_issued,
    COALESCE(SUM(cs.total_amount_used), 0) as total_amount_used,
    COALESCE(SUM(cs.remaining_balance), 0) as total_remaining
FROM subsidies.providers p
LEFT JOIN subsidies.customer_subsidies cs ON cs.provider_id = p.id
WHERE p.id = $1
GROUP BY p.id, p.name;
