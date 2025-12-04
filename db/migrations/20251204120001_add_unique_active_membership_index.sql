-- +goose Up
-- +goose StatementBegin

-- Add unique partial index to prevent duplicate active memberships for the same plan
-- This ensures a customer can only have ONE active membership per plan at a time
-- The partial index only applies where status = 'active', so multiple canceled/expired records are fine
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_active_membership_per_plan
ON users.customer_membership_plans (customer_id, membership_plan_id)
WHERE status = 'active';

-- Add index on stripe_subscription_id for faster lookups in webhook handlers
CREATE INDEX IF NOT EXISTS idx_customer_membership_stripe_sub_id
ON users.customer_membership_plans (stripe_subscription_id)
WHERE stripe_subscription_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS users.idx_customer_membership_stripe_sub_id;
DROP INDEX IF EXISTS users.idx_unique_active_membership_per_plan;

-- +goose StatementEnd
