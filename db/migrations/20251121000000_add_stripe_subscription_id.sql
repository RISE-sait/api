-- +goose Up
-- +goose StatementBegin

-- Migration: Add stripe_subscription_id to customer_membership_plans
-- This allows tracking which specific Stripe subscription is associated with each membership
-- Related to: Stripe webhook handling and reconciliation job improvements

-- Add stripe_subscription_id column
ALTER TABLE users.customer_membership_plans
ADD COLUMN IF NOT EXISTS stripe_subscription_id VARCHAR(255);

-- Add index for stripe_subscription_id to improve lookups
CREATE INDEX IF NOT EXISTS idx_customer_membership_stripe_subscription
ON users.customer_membership_plans(stripe_subscription_id)
WHERE stripe_subscription_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Rollback: Remove stripe_subscription_id column and index
DROP INDEX IF EXISTS idx_customer_membership_stripe_subscription;

ALTER TABLE users.customer_membership_plans
DROP COLUMN IF EXISTS stripe_subscription_id;

-- +goose StatementEnd
