-- +goose Up
ALTER TABLE users.customer_membership_plans
    ADD COLUMN IF NOT EXISTS last_stripe_event_at TIMESTAMPTZ;

COMMENT ON COLUMN users.customer_membership_plans.last_stripe_event_at
    IS 'Timestamp of the most recently processed Stripe event for this subscription. Used to reject out-of-order webhook events.';

-- +goose Down
ALTER TABLE users.customer_membership_plans
    DROP COLUMN IF EXISTS last_stripe_event_at;
