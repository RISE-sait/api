-- +goose Up
-- +goose StatementBegin
-- Create enum types
CREATE TYPE IF NOT EXISTS discount_duration_type AS ENUM ('once', 'repeating', 'forever');
CREATE TYPE IF NOT EXISTS discount_type AS ENUM ('percentage', 'fixed_amount');
CREATE TYPE IF NOT EXISTS discount_applies_to AS ENUM ('subscription', 'one_time', 'both');

-- Add new columns to discounts table
ALTER TABLE discounts ADD COLUMN IF NOT EXISTS stripe_coupon_id VARCHAR(255);
ALTER TABLE discounts ADD COLUMN IF NOT EXISTS duration_type discount_duration_type NOT NULL DEFAULT 'once';
ALTER TABLE discounts ADD COLUMN IF NOT EXISTS duration_months INT;
ALTER TABLE discounts ADD COLUMN IF NOT EXISTS discount_type discount_type NOT NULL DEFAULT 'percentage';
ALTER TABLE discounts ADD COLUMN IF NOT EXISTS discount_amount DECIMAL(10, 2);
ALTER TABLE discounts ADD COLUMN IF NOT EXISTS applies_to discount_applies_to NOT NULL DEFAULT 'both';
ALTER TABLE discounts ADD COLUMN IF NOT EXISTS max_redemptions INT;
ALTER TABLE discounts ADD COLUMN IF NOT EXISTS times_redeemed INT NOT NULL DEFAULT 0;

-- Add unique constraint on stripe_coupon_id
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'discounts_stripe_coupon_id_key'
    ) THEN
        ALTER TABLE discounts ADD CONSTRAINT discounts_stripe_coupon_id_key UNIQUE (stripe_coupon_id);
    END IF;
END
$$;

-- Add check constraints
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'check_duration_months'
    ) THEN
        ALTER TABLE discounts ADD CONSTRAINT check_duration_months CHECK (
            (duration_type = 'repeating' AND duration_months IS NOT NULL AND duration_months > 0) OR
            (duration_type != 'repeating' AND duration_months IS NULL)
        );
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'check_discount_value'
    ) THEN
        ALTER TABLE discounts ADD CONSTRAINT check_discount_value CHECK (
            (discount_type = 'percentage' AND discount_percent IS NOT NULL AND discount_percent > 0 AND discount_percent <= 100) OR
            (discount_type = 'fixed_amount' AND discount_amount IS NOT NULL AND discount_amount > 0)
        );
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'check_max_redemptions'
    ) THEN
        ALTER TABLE discounts ADD CONSTRAINT check_max_redemptions CHECK (
            max_redemptions IS NULL OR max_redemptions > 0
        );
    END IF;
END
$$;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_discounts_stripe_coupon_id ON discounts(stripe_coupon_id);
CREATE INDEX IF NOT EXISTS idx_discounts_active ON discounts(is_active) WHERE is_active = true;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop indexes
DROP INDEX IF EXISTS idx_discounts_stripe_coupon_id;
DROP INDEX IF EXISTS idx_discounts_active;

-- Remove constraints
ALTER TABLE discounts DROP CONSTRAINT IF EXISTS check_duration_months;
ALTER TABLE discounts DROP CONSTRAINT IF EXISTS check_discount_value;
ALTER TABLE discounts DROP CONSTRAINT IF EXISTS check_max_redemptions;
ALTER TABLE discounts DROP CONSTRAINT IF EXISTS discounts_stripe_coupon_id_key;

-- Remove columns
ALTER TABLE discounts DROP COLUMN IF EXISTS stripe_coupon_id;
ALTER TABLE discounts DROP COLUMN IF EXISTS duration_type;
ALTER TABLE discounts DROP COLUMN IF EXISTS duration_months;
ALTER TABLE discounts DROP COLUMN IF EXISTS discount_type;
ALTER TABLE discounts DROP COLUMN IF EXISTS discount_amount;
ALTER TABLE discounts DROP COLUMN IF EXISTS applies_to;
ALTER TABLE discounts DROP COLUMN IF EXISTS max_redemptions;
ALTER TABLE discounts DROP COLUMN IF EXISTS times_redeemed;

-- Drop enum types
DROP TYPE IF EXISTS discount_duration_type;
DROP TYPE IF EXISTS discount_type;
DROP TYPE IF EXISTS discount_applies_to;
-- +goose StatementEnd
