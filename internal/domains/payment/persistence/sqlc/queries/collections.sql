-- name: CreateCollectionAttempt :one
INSERT INTO payments.collection_attempts (
    customer_id,
    admin_id,
    amount_attempted,
    amount_collected,
    collection_method,
    payment_method_details,
    status,
    failure_reason,
    stripe_payment_intent_id,
    stripe_payment_link_id,
    stripe_customer_id,
    membership_plan_id,
    stripe_subscription_id,
    notes,
    previous_balance,
    new_balance,
    completed_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
) RETURNING *;

-- name: GetCollectionAttempt :one
SELECT * FROM payments.collection_attempts
WHERE id = $1;

-- name: UpdateCollectionAttemptStatus :one
UPDATE payments.collection_attempts
SET status = $2,
    amount_collected = COALESCE($3, amount_collected),
    failure_reason = $4,
    new_balance = $5,
    completed_at = CASE WHEN $2 IN ('success', 'failed') THEN CURRENT_TIMESTAMP ELSE completed_at END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ListCollectionAttempts :many
SELECT * FROM payments.collection_attempts
WHERE
    (sqlc.narg('customer_id')::uuid IS NULL OR customer_id = sqlc.narg('customer_id')) AND
    (sqlc.narg('admin_id')::uuid IS NULL OR admin_id = sqlc.narg('admin_id')) AND
    (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')) AND
    (sqlc.narg('collection_method')::text IS NULL OR collection_method = sqlc.narg('collection_method')) AND
    (sqlc.narg('start_date')::timestamptz IS NULL OR created_at >= sqlc.narg('start_date')) AND
    (sqlc.narg('end_date')::timestamptz IS NULL OR created_at <= sqlc.narg('end_date'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountCollectionAttempts :one
SELECT COUNT(*) FROM payments.collection_attempts
WHERE
    (sqlc.narg('customer_id')::uuid IS NULL OR customer_id = sqlc.narg('customer_id')) AND
    (sqlc.narg('admin_id')::uuid IS NULL OR admin_id = sqlc.narg('admin_id')) AND
    (sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')) AND
    (sqlc.narg('collection_method')::text IS NULL OR collection_method = sqlc.narg('collection_method')) AND
    (sqlc.narg('start_date')::timestamptz IS NULL OR created_at >= sqlc.narg('start_date')) AND
    (sqlc.narg('end_date')::timestamptz IS NULL OR created_at <= sqlc.narg('end_date'));

-- name: GetCollectionAttemptsByCustomer :many
SELECT * FROM payments.collection_attempts
WHERE customer_id = $1
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: GetCollectionSummary :one
SELECT
    COUNT(*) as total_attempts,
    COUNT(*) FILTER (WHERE status = 'success') as successful_attempts,
    COUNT(*) FILTER (WHERE status = 'failed') as failed_attempts,
    COUNT(*) FILTER (WHERE status = 'pending') as pending_attempts,
    COALESCE(SUM(amount_attempted), 0) as total_amount_attempted,
    COALESCE(SUM(amount_collected), 0) as total_amount_collected
FROM payments.collection_attempts
WHERE
    (sqlc.narg('start_date')::timestamptz IS NULL OR created_at >= sqlc.narg('start_date')) AND
    (sqlc.narg('end_date')::timestamptz IS NULL OR created_at <= sqlc.narg('end_date'));

-- Payment Links queries

-- name: CreatePaymentLink :one
INSERT INTO payments.payment_links (
    customer_id,
    admin_id,
    stripe_payment_link_id,
    stripe_payment_link_url,
    amount,
    description,
    membership_plan_id,
    collection_attempt_id,
    status,
    sent_via,
    sent_to_email,
    sent_to_phone,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
) RETURNING *;

-- name: GetPaymentLink :one
SELECT * FROM payments.payment_links
WHERE id = $1;

-- name: GetPaymentLinkByStripeID :one
SELECT * FROM payments.payment_links
WHERE stripe_payment_link_id = $1;

-- name: UpdatePaymentLinkStatus :one
UPDATE payments.payment_links
SET status = $2,
    sent_at = CASE WHEN $2 = 'sent' AND sent_at IS NULL THEN CURRENT_TIMESTAMP ELSE sent_at END,
    opened_at = CASE WHEN $2 = 'opened' AND opened_at IS NULL THEN CURRENT_TIMESTAMP ELSE opened_at END,
    completed_at = CASE WHEN $2 = 'completed' AND completed_at IS NULL THEN CURRENT_TIMESTAMP ELSE completed_at END
WHERE id = $1
RETURNING *;

-- name: ListPaymentLinksByCustomer :many
SELECT * FROM payments.payment_links
WHERE customer_id = $1
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListPendingPaymentLinks :many
SELECT * FROM payments.payment_links
WHERE status IN ('pending', 'sent')
    AND (expires_at IS NULL OR expires_at > CURRENT_TIMESTAMP)
ORDER BY created_at DESC;

-- name: ExpireOldPaymentLinks :exec
UPDATE payments.payment_links
SET status = 'expired'
WHERE status IN ('pending', 'sent')
    AND expires_at IS NOT NULL
    AND expires_at <= CURRENT_TIMESTAMP;
