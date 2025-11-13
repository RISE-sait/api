-- name: CreateCustomerSubsidy :one
INSERT INTO subsidies.customer_subsidies (
    customer_id,
    provider_id,
    approved_amount,
    status,
    approved_by,
    approved_at,
    valid_from,
    valid_until,
    reason,
    admin_notes
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetCustomerSubsidy :one
SELECT
    cs.*,
    p.name as provider_name,
    u.first_name || ' ' || u.last_name as customer_name,
    u.email as customer_email
FROM subsidies.customer_subsidies cs
LEFT JOIN subsidies.providers p ON p.id = cs.provider_id
LEFT JOIN users.users u ON u.id = cs.customer_id
WHERE cs.id = $1;

-- name: GetActiveSubsidyForCustomer :one
SELECT
    cs.*,
    p.name as provider_name
FROM subsidies.customer_subsidies cs
LEFT JOIN subsidies.providers p ON p.id = cs.provider_id
WHERE cs.customer_id = $1
  AND cs.status IN ('approved', 'active')
  AND cs.remaining_balance > 0
  AND cs.valid_from <= CURRENT_TIMESTAMP
  AND (cs.valid_until IS NULL OR cs.valid_until >= CURRENT_TIMESTAMP)
ORDER BY cs.created_at ASC
LIMIT 1;

-- name: ListCustomerSubsidies :many
SELECT
    cs.*,
    p.name as provider_name,
    u.first_name || ' ' || u.last_name as customer_name,
    u.email as customer_email
FROM subsidies.customer_subsidies cs
LEFT JOIN subsidies.providers p ON p.id = cs.provider_id
LEFT JOIN users.users u ON u.id = cs.customer_id
WHERE (sqlc.narg('customer_id')::uuid IS NULL OR cs.customer_id = sqlc.narg('customer_id'))
  AND (sqlc.narg('provider_id')::uuid IS NULL OR cs.provider_id = sqlc.narg('provider_id'))
  AND (sqlc.narg('status')::text IS NULL OR cs.status = sqlc.narg('status'))
ORDER BY cs.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountCustomerSubsidies :one
SELECT COUNT(*) FROM subsidies.customer_subsidies cs
WHERE (sqlc.narg('customer_id')::uuid IS NULL OR cs.customer_id = sqlc.narg('customer_id'))
  AND (sqlc.narg('provider_id')::uuid IS NULL OR cs.provider_id = sqlc.narg('provider_id'))
  AND (sqlc.narg('status')::text IS NULL OR cs.status = sqlc.narg('status'));

-- name: GetCustomerSubsidiesByCustomerID :many
SELECT
    cs.*,
    p.name as provider_name
FROM subsidies.customer_subsidies cs
LEFT JOIN subsidies.providers p ON p.id = cs.provider_id
WHERE cs.customer_id = $1
ORDER BY cs.created_at DESC;

-- name: UpdateSubsidyStatus :one
UPDATE subsidies.customer_subsidies
SET
    status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateSubsidyUsage :one
UPDATE subsidies.customer_subsidies
SET
    total_amount_used = total_amount_used + $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: DeactivateSubsidy :one
UPDATE subsidies.customer_subsidies
SET
    status = 'expired',
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: MarkSubsidyAsDepleted :one
UPDATE subsidies.customer_subsidies
SET
    status = 'depleted',
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ExpireExpiredSubsidies :exec
UPDATE subsidies.customer_subsidies
SET
    status = 'expired',
    updated_at = CURRENT_TIMESTAMP
WHERE status IN ('approved', 'active')
  AND valid_until IS NOT NULL
  AND valid_until < CURRENT_TIMESTAMP;

-- name: GetSubsidySummary :one
SELECT
    COUNT(*) FILTER (WHERE status = 'active') as active_count,
    COUNT(*) FILTER (WHERE status = 'pending') as pending_count,
    COUNT(*) FILTER (WHERE status = 'depleted') as depleted_count,
    COALESCE(SUM(approved_amount) FILTER (WHERE status IN ('approved', 'active')), 0) as total_approved,
    COALESCE(SUM(total_amount_used), 0) as total_used,
    COALESCE(SUM(remaining_balance) FILTER (WHERE status IN ('approved', 'active')), 0) as total_remaining
FROM subsidies.customer_subsidies;
