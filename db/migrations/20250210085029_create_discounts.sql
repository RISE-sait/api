-- +goose Up
-- +goose StatementBegin

-- Create enum types for discounts
CREATE TYPE discount_duration_type AS ENUM ('once', 'repeating', 'forever');
CREATE TYPE discount_type AS ENUM ('percentage', 'fixed_amount');
CREATE TYPE discount_applies_to AS ENUM ('subscription', 'one_time', 'both');

-- Create base discounts table
CREATE TABLE IF NOT EXISTS discounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    description TEXT,
    discount_percent INT NOT NULL DEFAULT 0,
    discount_type discount_type NOT NULL DEFAULT 'percentage',
    is_use_unlimited BOOLEAN NOT NULL DEFAULT FALSE,
    use_per_client INT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_to TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT check_use_per_client CHECK (
        (is_use_unlimited = true) OR
        (is_use_unlimited = false AND use_per_client IS NOT NULL)
    )
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS discounts;
DROP TYPE IF EXISTS discount_duration_type;
DROP TYPE IF EXISTS discount_type;
DROP TYPE IF EXISTS discount_applies_to;

-- +goose StatementEnd
