-- +goose Up

-- Add enum types for discount configuration (only if they don't exist)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'discount_duration_type') THEN
        CREATE TYPE discount_duration_type AS ENUM ('once', 'repeating', 'forever');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'discount_type') THEN
        CREATE TYPE discount_type AS ENUM ('percentage', 'fixed_amount');
    END IF;
END
$$;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'discount_applies_to') THEN
        CREATE TYPE discount_applies_to AS ENUM ('subscription', 'one_time', 'both');
    END IF;
END
$$;

-- Add new columns to discounts table for Stripe integration (only if they don't exist)
DO $$
BEGIN
    ALTER TABLE discounts ADD COLUMN IF NOT EXISTS stripe_coupon_id VARCHAR(255);
END
$$;

DO $$
BEGIN
    ALTER TABLE discounts ADD COLUMN IF NOT EXISTS duration_type discount_duration_type NOT NULL DEFAULT 'once'::discount_duration_type;
END
$$;

DO $$
BEGIN
    ALTER TABLE discounts ADD COLUMN IF NOT EXISTS duration_months INT;
END
$$;

DO $$
BEGIN
    ALTER TABLE discounts ADD COLUMN IF NOT EXISTS discount_type discount_type NOT NULL DEFAULT 'percentage'::discount_type;
END
$$;

DO $$
BEGIN
    ALTER TABLE discounts ADD COLUMN IF NOT EXISTS discount_amount DECIMAL(10, 2);
END
$$;

DO $$
BEGIN
    ALTER TABLE discounts ADD COLUMN IF NOT EXISTS applies_to discount_applies_to NOT NULL DEFAULT 'both'::discount_applies_to;
END
$$;

DO $$
BEGIN
    ALTER TABLE discounts ADD COLUMN IF NOT EXISTS max_redemptions INT;
END
$$;

DO $$
BEGIN
    ALTER TABLE discounts ADD COLUMN IF NOT EXISTS times_redeemed INT NOT NULL DEFAULT 0;
END
$$;

-- Add unique constraint on stripe_coupon_id (only if it doesn't exist)
DO $$
BEGIN
    ALTER TABLE discounts ADD CONSTRAINT discounts_stripe_coupon_id_key UNIQUE (stripe_coupon_id);
EXCEPTION
    WHEN duplicate_object THEN NULL;
END
$$;

-- Add constraints (only if they don't exist)
DO $$
BEGIN
    ALTER TABLE discounts
    ADD CONSTRAINT check_duration_months CHECK (
        (duration_type = 'repeating' AND duration_months IS NOT NULL AND duration_months > 0) OR
        (duration_type != 'repeating' AND duration_months IS NULL)
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END
$$;

DO $$
BEGIN
    ALTER TABLE discounts
    ADD CONSTRAINT check_discount_value CHECK (
        (discount_type = 'percentage' AND discount_percent IS NOT NULL AND discount_percent > 0 AND discount_percent <= 100) OR
        (discount_type = 'fixed_amount' AND discount_amount IS NOT NULL AND discount_amount > 0)
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END
$$;

DO $$
BEGIN
    ALTER TABLE discounts
    ADD CONSTRAINT check_max_redemptions CHECK (
        max_redemptions IS NULL OR max_redemptions > 0
    );
EXCEPTION
    WHEN duplicate_object THEN NULL;
END
$$;

-- Create indexes (only if they don't exist)
CREATE INDEX IF NOT EXISTS idx_discounts_stripe_coupon_id ON discounts(stripe_coupon_id);
CREATE INDEX IF NOT EXISTS idx_discounts_active ON discounts(is_active) WHERE is_active = true;

-- +goose Down

-- Drop indexes
DROP INDEX IF EXISTS idx_discounts_stripe_coupon_id;
DROP INDEX IF EXISTS idx_discounts_active;

-- Remove constraints
ALTER TABLE discounts
DROP CONSTRAINT IF EXISTS check_duration_months,
DROP CONSTRAINT IF EXISTS check_discount_value,
DROP CONSTRAINT IF EXISTS check_max_redemptions,
DROP CONSTRAINT IF EXISTS discounts_stripe_coupon_id_key;

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
