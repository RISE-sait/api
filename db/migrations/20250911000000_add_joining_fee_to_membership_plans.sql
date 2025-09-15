-- +goose Up
-- +goose StatementBegin

-- Add joining_fee column to membership_plans table
ALTER TABLE membership.membership_plans 
ADD COLUMN joining_fee INTEGER DEFAULT 0 NOT NULL;

-- Add comment to explain the field
COMMENT ON COLUMN membership.membership_plans.joining_fee IS 'One-time joining fee in cents (e.g., 13000 = $130.00). Applied as Stripe setup fee on first payment only.';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove joining_fee column
ALTER TABLE membership.membership_plans 
DROP COLUMN IF EXISTS joining_fee;

-- +goose StatementEnd