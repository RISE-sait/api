-- +goose Up
-- Create payment schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS payment;

-- Webhook idempotency tracking (database-backed to survive restarts)
CREATE TABLE IF NOT EXISTS payment.webhook_events (
    event_id VARCHAR(255) PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    processed_at TIMESTAMPTZ DEFAULT NOW(),
    status VARCHAR(20) DEFAULT 'processed'
);

CREATE INDEX idx_webhook_events_processed_at ON payment.webhook_events(processed_at);
CREATE INDEX idx_webhook_events_event_type ON payment.webhook_events(event_type);

-- Failed refunds queue for recovery
CREATE TABLE IF NOT EXISTS payment.failed_refunds (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    event_id UUID REFERENCES events.events(id) ON DELETE SET NULL,
    credit_amount INT NOT NULL,
    error_message TEXT,
    retry_count INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending', -- pending, resolved, failed
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX idx_failed_refunds_status ON payment.failed_refunds(status);
CREATE INDEX idx_failed_refunds_customer ON payment.failed_refunds(customer_id);

-- Dead letter queue for failed webhooks
CREATE TABLE IF NOT EXISTS payment.failed_webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB,
    error_message TEXT,
    attempts INT DEFAULT 1,
    status VARCHAR(20) DEFAULT 'failed', -- failed, resolved, ignored
    created_at TIMESTAMPTZ DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX idx_failed_webhooks_status ON payment.failed_webhooks(status);
CREATE INDEX idx_failed_webhooks_event_type ON payment.failed_webhooks(event_type);

-- +goose Down
DROP TABLE IF EXISTS payment.failed_webhooks;
DROP TABLE IF EXISTS payment.failed_refunds;
DROP TABLE IF EXISTS payment.webhook_events;
-- Note: Not dropping payment schema as other tables may use it
