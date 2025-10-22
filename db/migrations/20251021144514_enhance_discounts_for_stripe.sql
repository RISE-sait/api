-- +goose Up
-- +goose StatementBegin

-- Add enum types for discount configuration
CREATE TYPE discount_duration_type AS ENUM ('once', 'repeating', 'forever');
CREATE TYPE discount_type AS ENUM ('percentage', 'fixed_amount');
CREATE TYPE discount_applies_to AS ENUM ('subscription', 'one_time', 'both');

-- Add new columns to discounts table for Stripe integration
ALTER TABLE discounts
ADD COLUMN stripe_coupon_id VARCHAR(255) UNIQUE,
ADD COLUMN duration_type discount_duration_type NOT NULL DEFAULT 'once',
ADD COLUMN duration_months INT,
ADD COLUMN discount_type discount_type NOT NULL DEFAULT 'percentage',
ADD COLUMN discount_amount DECIMAL(10, 2),
ADD COLUMN applies_to discount_applies_to NOT NULL DEFAULT 'both',
ADD COLUMN max_redemptions INT,
ADD COLUMN times_redeemed INT NOT NULL DEFAULT 0;

-- Add constraints
ALTER TABLE discounts
ADD CONSTRAINT check_duration_months CHECK (
    (duration_type = 'repeating' AND duration_months IS NOT NULL AND duration_months > 0) OR
    (duration_type != 'repeating' AND duration_months IS NULL)
);

ALTER TABLE discounts
ADD CONSTRAINT check_discount_value CHECK (
    (discount_type = 'percentage' AND discount_percent IS NOT NULL AND discount_percent > 0 AND discount_percent <= 100) OR
    (discount_type = 'fixed_amount' AND discount_amount IS NOT NULL AND discount_amount > 0)
);

ALTER TABLE discounts
ADD CONSTRAINT check_max_redemptions CHECK (
    max_redemptions IS NULL OR max_redemptions > 0
);

-- Create index on stripe_coupon_id for faster lookups
CREATE INDEX idx_discounts_stripe_coupon_id ON discounts(stripe_coupon_id);

-- Create index on code lookups (assuming we'll add a code field later)
CREATE INDEX idx_discounts_active ON discounts(is_active) WHERE is_active = true;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop indexes
DROP INDEX IF EXISTS idx_discounts_stripe_coupon_id;
DROP INDEX IF EXISTS idx_discounts_active;

-- Remove constraints
ALTER TABLE discounts
DROP CONSTRAINT IF EXISTS check_duration_months,
DROP CONSTRAINT IF EXISTS check_discount_value,
DROP CONSTRAINT IF EXISTS check_max_redemptions;

-- Remove columns
ALTER TABLE discounts
DROP COLUMN IF EXISTS stripe_coupon_id,
DROP COLUMN IF EXISTS duration_type,
DROP COLUMN IF EXISTS duration_months,
DROP COLUMN IF EXISTS discount_type,
DROP COLUMN IF EXISTS discount_amount,
DROP COLUMN IF EXISTS applies_to,
DROP COLUMN IF EXISTS max_redemptions,
DROP COLUMN IF EXISTS times_redeemed;

-- Drop enum types
DROP TYPE IF EXISTS discount_duration_type;
DROP TYPE IF EXISTS discount_type;
DROP TYPE IF EXISTS discount_applies_to;

-- +goose StatementEnd
