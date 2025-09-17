-- +goose Up
-- +goose StatementBegin

-- Add optional credit allocation and weekly limit fields to membership plans
-- These are only used for credit-based memberships, NULL for traditional memberships
ALTER TABLE membership.membership_plans 
ADD COLUMN credit_allocation INTEGER NULL,
ADD COLUMN weekly_credit_limit INTEGER NULL;

-- Add comments to explain the fields
COMMENT ON COLUMN membership.membership_plans.credit_allocation IS 'Number of credits awarded when purchasing this membership plan (NULL for non-credit memberships)';
COMMENT ON COLUMN membership.membership_plans.weekly_credit_limit IS 'Maximum credits that can be used per week with this membership plan (NULL for non-credit memberships, 0 = unlimited credits)';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove the credit fields
ALTER TABLE membership.membership_plans 
DROP COLUMN IF EXISTS credit_allocation,
DROP COLUMN IF EXISTS weekly_credit_limit;

-- +goose StatementEnd