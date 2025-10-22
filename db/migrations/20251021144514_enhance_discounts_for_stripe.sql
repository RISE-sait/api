-- +goose Up
-- +goose StatementBegin
-- Create enum types
CREATE TYPE IF NOT EXISTS discount_duration_type AS ENUM ('once', 'repeating', 'forever');
CREATE TYPE IF NOT EXISTS discount_type AS ENUM ('percentage', 'fixed_amount');
CREATE TYPE IF NOT EXISTS discount_applies_to AS ENUM ('subscription', 'one_time', 'both');

-- Create discounts table with all columns (handles both fresh DB and existing table)
CREATE TABLE IF NOT EXISTS discounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    description TEXT,
    discount_percent INT NOT NULL DEFAULT 0,
    discount_amount DECIMAL(10, 2),
    discount_type discount_type NOT NULL DEFAULT 'percentage',
    is_use_unlimited BOOLEAN NOT NULL DEFAULT FALSE,
    use_per_client INT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_to TIMESTAMP WITH TIME ZONE,
    duration_type discount_duration_type NOT NULL DEFAULT 'once',
    duration_months INT,
    applies_to discount_applies_to NOT NULL DEFAULT 'both',
    max_redemptions INT,
    times_redeemed INT NOT NULL DEFAULT 0,
    stripe_coupon_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT check_use_per_client CHECK (
        (is_use_unlimited = true) OR
        (is_use_unlimited = false AND use_per_client IS NOT NULL)
    ),
    CONSTRAINT check_duration_months CHECK (
        (duration_type = 'repeating' AND duration_months IS NOT NULL AND duration_months > 0) OR
        (duration_type != 'repeating' AND duration_months IS NULL)
    ),
    CONSTRAINT check_discount_value CHECK (
        (discount_type = 'percentage' AND discount_percent IS NOT NULL AND discount_percent > 0 AND discount_percent <= 100) OR
        (discount_type = 'fixed_amount' AND discount_amount IS NOT NULL AND discount_amount > 0)
    ),
    CONSTRAINT check_max_redemptions CHECK (
        max_redemptions IS NULL OR max_redemptions > 0
    ),
    CONSTRAINT discounts_stripe_coupon_id_key UNIQUE (stripe_coupon_id)
);

-- Add columns if table already exists (for local DBs that already have the old schema)
DO $$
BEGIN
    ALTER TABLE discounts ADD COLUMN IF NOT EXISTS stripe_coupon_id VARCHAR(255);
    ALTER TABLE discounts ADD COLUMN IF NOT EXISTS duration_type discount_duration_type NOT NULL DEFAULT 'once';
    ALTER TABLE discounts ADD COLUMN IF NOT EXISTS duration_months INT;
    ALTER TABLE discounts ADD COLUMN IF NOT EXISTS discount_type discount_type NOT NULL DEFAULT 'percentage';
    ALTER TABLE discounts ADD COLUMN IF NOT EXISTS discount_amount DECIMAL(10, 2);
    ALTER TABLE discounts ADD COLUMN IF NOT EXISTS applies_to discount_applies_to NOT NULL DEFAULT 'both';
    ALTER TABLE discounts ADD COLUMN IF NOT EXISTS max_redemptions INT;
    ALTER TABLE discounts ADD COLUMN IF NOT EXISTS times_redeemed INT NOT NULL DEFAULT 0;
EXCEPTION
    WHEN duplicate_column THEN NULL;
END
$$;

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

-- Drop the entire table
DROP TABLE IF EXISTS discounts;

-- Drop enum types
DROP TYPE IF EXISTS discount_duration_type;
DROP TYPE IF EXISTS discount_type;
DROP TYPE IF EXISTS discount_applies_to;
-- +goose StatementEnd
