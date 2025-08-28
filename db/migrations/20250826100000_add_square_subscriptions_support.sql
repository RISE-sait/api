-- +goose Up
-- +goose StatementBegin

-- Add Square customer ID to users table
ALTER TABLE users.users 
ADD COLUMN square_customer_id VARCHAR(255);

-- Add Square subscription tracking to customer membership plans
ALTER TABLE users.customer_membership_plans 
ADD COLUMN square_subscription_id VARCHAR(255),
ADD COLUMN subscription_status VARCHAR(50),
ADD COLUMN next_billing_date TIMESTAMPTZ,
ADD COLUMN subscription_created_at TIMESTAMPTZ,
ADD COLUMN subscription_source VARCHAR(20) DEFAULT 'one-time' CHECK (subscription_source IN ('one-time', 'subscription'));

-- Index for Square customer lookups
CREATE INDEX idx_users_square_customer 
ON users.users(square_customer_id);

-- Index for Square subscription lookups  
CREATE INDEX idx_customer_membership_square_subscription 
ON users.customer_membership_plans(square_subscription_id);

-- Index for subscription status queries
CREATE INDEX idx_customer_membership_subscription_status 
ON users.customer_membership_plans(subscription_status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove indexes
DROP INDEX IF EXISTS idx_customer_membership_subscription_status;
DROP INDEX IF EXISTS idx_customer_membership_square_subscription;
DROP INDEX IF EXISTS idx_users_square_customer;

-- Remove columns
ALTER TABLE users.customer_membership_plans 
DROP COLUMN IF EXISTS subscription_source,
DROP COLUMN IF EXISTS subscription_created_at,
DROP COLUMN IF EXISTS next_billing_date,
DROP COLUMN IF EXISTS subscription_status,
DROP COLUMN IF EXISTS square_subscription_id;

ALTER TABLE users.users 
DROP COLUMN IF EXISTS square_customer_id;

-- +goose StatementEnd