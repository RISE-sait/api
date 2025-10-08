-- +goose Up
-- +goose StatementBegin

-- Create credit_packages table for one-time credit purchases
CREATE TABLE IF NOT EXISTS users.credit_packages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(150) NOT NULL,
    description TEXT,
    stripe_price_id VARCHAR(50) NOT NULL UNIQUE,
    credit_allocation INTEGER NOT NULL CHECK (credit_allocation > 0),
    weekly_credit_limit INTEGER NOT NULL CHECK (weekly_credit_limit >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add comments
COMMENT ON TABLE users.credit_packages IS 'Available credit packages for one-time purchase';
COMMENT ON COLUMN users.credit_packages.credit_allocation IS 'Number of credits awarded when purchasing this package';
COMMENT ON COLUMN users.credit_packages.weekly_credit_limit IS 'Maximum credits that can be used per week (0 = unlimited)';
COMMENT ON COLUMN users.credit_packages.stripe_price_id IS 'Stripe price ID for one-time payment checkout';

-- Add index for Stripe price lookups
CREATE INDEX IF NOT EXISTS idx_credit_packages_stripe_price_id ON users.credit_packages(stripe_price_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_credit_packages_stripe_price_id;
DROP TABLE IF EXISTS users.credit_packages;

-- +goose StatementEnd
