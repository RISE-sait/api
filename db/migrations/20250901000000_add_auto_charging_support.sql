-- +goose Up
-- +goose StatementBegin

-- Create separate table for auto-charging status tracking
CREATE TABLE users.subscription_auto_charging (
    id uuid NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_membership_plan_id uuid NOT NULL REFERENCES users.customer_membership_plans(id) ON DELETE CASCADE,
    square_subscription_id VARCHAR(255),
    enabled BOOLEAN DEFAULT false,
    card_id VARCHAR(255),
    last_payment_id VARCHAR(255),
    error_type VARCHAR(100),
    error_details TEXT,
    retry_count INTEGER DEFAULT 0,
    permanently_failed BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for efficient queries
CREATE UNIQUE INDEX idx_subscription_auto_charging_membership_plan 
ON users.subscription_auto_charging(customer_membership_plan_id);

CREATE INDEX idx_subscription_auto_charging_square_id 
ON users.subscription_auto_charging(square_subscription_id);

CREATE INDEX idx_subscription_auto_charging_failed 
ON users.subscription_auto_charging(enabled, updated_at) 
WHERE enabled = false AND error_type IS NOT NULL;

CREATE INDEX idx_subscription_auto_charging_retry 
ON users.subscription_auto_charging(retry_count, permanently_failed, updated_at)
WHERE enabled = false AND error_type IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove table and indexes
DROP TABLE IF EXISTS users.subscription_auto_charging;

-- +goose StatementEnd