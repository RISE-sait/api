-- name: CreateAuditLog :one
INSERT INTO subsidies.audit_log (
    customer_subsidy_id,
    action,
    performed_by,
    previous_status,
    new_status,
    amount_changed,
    notes,
    ip_address
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetAuditLog :one
SELECT
    al.*,
    u.first_name || ' ' || u.last_name as performed_by_name
FROM subsidies.audit_log al
LEFT JOIN users.users u ON u.id = al.performed_by
WHERE al.id = $1;

-- name: ListAuditLogsBySubsidy :many
SELECT
    al.*,
    u.first_name || ' ' || u.last_name as performed_by_name
FROM subsidies.audit_log al
LEFT JOIN users.users u ON u.id = al.performed_by
WHERE al.customer_subsidy_id = $1
ORDER BY al.created_at DESC;

-- name: ListAuditLogs :many
SELECT
    al.*,
    u.first_name || ' ' || u.last_name as performed_by_name,
    cs.customer_id
FROM subsidies.audit_log al
LEFT JOIN users.users u ON u.id = al.performed_by
LEFT JOIN subsidies.customer_subsidies cs ON cs.id = al.customer_subsidy_id
WHERE (sqlc.narg('subsidy_id')::uuid IS NULL OR al.customer_subsidy_id = sqlc.narg('subsidy_id'))
  AND (sqlc.narg('action')::text IS NULL OR al.action = sqlc.narg('action'))
ORDER BY al.created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountAuditLogs :one
SELECT COUNT(*) FROM subsidies.audit_log
WHERE (sqlc.narg('subsidy_id')::uuid IS NULL OR customer_subsidy_id = sqlc.narg('subsidy_id'))
  AND (sqlc.narg('action')::text IS NULL OR action = sqlc.narg('action'));
