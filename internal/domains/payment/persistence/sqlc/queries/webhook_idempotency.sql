-- name: CheckWebhookProcessed :one
SELECT EXISTS(
    SELECT 1 FROM payment.webhook_events
    WHERE event_id = $1
    AND processed_at > NOW() - INTERVAL '24 hours'
) AS is_processed;

-- name: MarkWebhookProcessed :exec
INSERT INTO payment.webhook_events (event_id, event_type, processed_at, status)
VALUES ($1, $2, NOW(), 'processed')
ON CONFLICT (event_id) DO NOTHING;

-- name: CleanupOldWebhookEvents :exec
DELETE FROM payment.webhook_events
WHERE processed_at < NOW() - INTERVAL '7 days';

-- name: InsertFailedRefund :one
INSERT INTO payment.failed_refunds (customer_id, event_id, credit_amount, error_message, status)
VALUES ($1, $2, $3, $4, 'pending')
RETURNING *;

-- name: GetPendingFailedRefunds :many
SELECT * FROM payment.failed_refunds
WHERE status = 'pending' AND retry_count < 5
ORDER BY created_at ASC
LIMIT $1;

-- name: UpdateFailedRefundRetry :exec
UPDATE payment.failed_refunds
SET retry_count = retry_count + 1, updated_at = NOW()
WHERE id = $1;

-- name: ResolveFailedRefund :exec
UPDATE payment.failed_refunds
SET status = 'resolved', resolved_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: MarkFailedRefundFailed :exec
UPDATE payment.failed_refunds
SET status = 'failed', updated_at = NOW()
WHERE id = $1;

-- name: InsertFailedWebhook :one
INSERT INTO payment.failed_webhooks (event_id, event_type, payload, error_message, attempts)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetFailedWebhooks :many
SELECT * FROM payment.failed_webhooks
WHERE status = 'failed'
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ResolveFailedWebhook :exec
UPDATE payment.failed_webhooks
SET status = 'resolved', resolved_at = NOW()
WHERE id = $1;

-- name: IgnoreFailedWebhook :exec
UPDATE payment.failed_webhooks
SET status = 'ignored', resolved_at = NOW()
WHERE id = $1;
