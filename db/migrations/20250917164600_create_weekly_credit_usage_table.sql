-- +goose Up
-- +goose StatementBegin

-- Create table to track weekly credit usage per customer
CREATE TABLE IF NOT EXISTS users.weekly_credit_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    week_start_date DATE NOT NULL, -- Monday of the week (ISO week)
    credits_used INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Ensure one record per customer per week
    CONSTRAINT unique_customer_week UNIQUE (customer_id, week_start_date)
);

-- Add indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_weekly_credit_usage_customer_id ON users.weekly_credit_usage (customer_id);
CREATE INDEX IF NOT EXISTS idx_weekly_credit_usage_week_start ON users.weekly_credit_usage (week_start_date);
CREATE INDEX IF NOT EXISTS idx_weekly_credit_usage_customer_week ON users.weekly_credit_usage (customer_id, week_start_date);

-- Add comments
COMMENT ON TABLE users.weekly_credit_usage IS 'Tracks weekly credit consumption per customer for membership limit enforcement';
COMMENT ON COLUMN users.weekly_credit_usage.week_start_date IS 'Monday of the ISO week (e.g., 2024-01-15 for week starting Jan 15)';
COMMENT ON COLUMN users.weekly_credit_usage.credits_used IS 'Total credits consumed during this week';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove indexes
DROP INDEX IF EXISTS idx_weekly_credit_usage_customer_id;
DROP INDEX IF EXISTS idx_weekly_credit_usage_week_start;
DROP INDEX IF EXISTS idx_weekly_credit_usage_customer_week;

-- Remove table
DROP TABLE IF EXISTS users.weekly_credit_usage;

-- +goose StatementEnd