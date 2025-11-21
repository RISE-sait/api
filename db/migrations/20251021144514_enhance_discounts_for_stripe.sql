-- +goose Up
-- +goose StatementBegin
-- Enhance discounts table with Stripe integration columns
-- Base table and enums already created by 20250210085029_create_discounts.sql

-- Add new columns for Stripe integration
ALTER TABLE discounts ADD COLUMN IF NOT EXISTS stripe_coupon_id VARCHAR(255);
ALTER TABLE discounts ADD COLUMN IF NOT EXISTS duration_type discount_duration_type NOT NULL DEFAULT 'once';
ALTER TABLE discounts ADD COLUMN IF NOT EXISTS duration_months INT;
ALTER TABLE discounts ADD COLUMN IF NOT EXISTS discount_amount DECIMAL(10, 2);
ALTER TABLE discounts ADD COLUMN IF NOT EXISTS applies_to discount_applies_to NOT NULL DEFAULT 'both';
ALTER TABLE discounts ADD COLUMN IF NOT EXISTS max_redemptions INT;
ALTER TABLE discounts ADD COLUMN IF NOT EXISTS times_redeemed INT NOT NULL DEFAULT 0;

-- Add constraints if they don't exist (for local DBs)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'discounts_stripe_coupon_id_key'
    ) THEN
        ALTER TABLE discounts ADD CONSTRAINT discounts_stripe_coupon_id_key UNIQUE (stripe_coupon_id);
    END IF;
END
$$;

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

-- Remove Stripe-specific columns (keep base table for earlier migrations)
ALTER TABLE discounts DROP COLUMN IF EXISTS stripe_coupon_id;
ALTER TABLE discounts DROP COLUMN IF EXISTS duration_type;
ALTER TABLE discounts DROP COLUMN IF EXISTS duration_months;
ALTER TABLE discounts DROP COLUMN IF EXISTS discount_amount;
ALTER TABLE discounts DROP COLUMN IF EXISTS applies_to;
ALTER TABLE discounts DROP COLUMN IF EXISTS max_redemptions;
ALTER TABLE discounts DROP COLUMN IF EXISTS times_redeemed;

-- Note: Base table and enum types remain (managed by 20250210085029_create_discounts.sql)
-- +goose StatementEnd
