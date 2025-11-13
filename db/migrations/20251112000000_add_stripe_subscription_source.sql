-- +goose Up
-- +goose StatementBegin

-- Migration: Add 'stripe' to subscription_source constraint
-- This allows memberships created via Stripe Checkout to be properly tracked
-- Related to: subscription_source bug fix and reconciliation jobs

-- Add 'stripe' to the subscription_source constraint
ALTER TABLE users.customer_membership_plans
DROP CONSTRAINT IF EXISTS customer_membership_plans_subscription_source_check;

ALTER TABLE users.customer_membership_plans
ADD CONSTRAINT customer_membership_plans_subscription_source_check
CHECK (subscription_source IN ('one-time', 'subscription', 'stripe'));

-- Update existing 'subscription' values to 'stripe' for Stripe-created memberships
-- This ensures consistency with the new code that uses 'stripe'
UPDATE users.customer_membership_plans
SET subscription_source = 'stripe'
WHERE subscription_source = 'subscription';

-- Add index for subscription_source to improve reconciliation job performance
CREATE INDEX IF NOT EXISTS idx_customer_membership_plans_subscription_source
ON users.customer_membership_plans(subscription_source)
WHERE subscription_source = 'stripe';

-- Add index for status + subscription_source (used by reconciliation job query)
CREATE INDEX IF NOT EXISTS idx_customer_membership_plans_status_source
ON users.customer_membership_plans(status, subscription_source)
WHERE subscription_source = 'stripe' AND status IN ('active', 'inactive');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Rollback: Remove 'stripe' from constraint and revert to original
DROP INDEX IF EXISTS idx_customer_membership_plans_status_source;
DROP INDEX IF EXISTS idx_customer_membership_plans_subscription_source;

-- Revert 'stripe' back to 'subscription'
UPDATE users.customer_membership_plans
SET subscription_source = 'subscription'
WHERE subscription_source = 'stripe';

-- Restore original constraint
ALTER TABLE users.customer_membership_plans
DROP CONSTRAINT IF EXISTS customer_membership_plans_subscription_source_check;

ALTER TABLE users.customer_membership_plans
ADD CONSTRAINT customer_membership_plans_subscription_source_check
CHECK (subscription_source IN ('one-time', 'subscription'));

-- +goose StatementEnd
