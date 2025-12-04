-- +goose Up
-- +goose StatementBegin

-- Add error_message column to webhook_events for tracking failed processing
ALTER TABLE payment.webhook_events
ADD COLUMN IF NOT EXISTS error_message TEXT;

-- Add index for status lookups (useful for retry/monitoring queries)
CREATE INDEX IF NOT EXISTS idx_webhook_events_status ON payment.webhook_events(status);

-- Add composite index for cleanup queries that check processed_at and status together
CREATE INDEX IF NOT EXISTS idx_webhook_events_cleanup ON payment.webhook_events(processed_at, status)
WHERE status = 'processing';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS payment.idx_webhook_events_cleanup;
DROP INDEX IF EXISTS payment.idx_webhook_events_status;

ALTER TABLE payment.webhook_events
DROP COLUMN IF EXISTS error_message;

-- +goose StatementEnd
