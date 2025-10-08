-- +goose Up
-- +goose StatementBegin

-- Create table to track each customer's active credit package
CREATE TABLE IF NOT EXISTS users.customer_active_credit_package (
    customer_id UUID PRIMARY KEY REFERENCES users.users(id) ON DELETE CASCADE,
    credit_package_id UUID NOT NULL REFERENCES users.credit_packages(id) ON DELETE RESTRICT,
    weekly_credit_limit INTEGER NOT NULL CHECK (weekly_credit_limit >= 0),
    purchased_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add comments
COMMENT ON TABLE users.customer_active_credit_package IS 'Tracks each customer''s currently active credit package and their weekly limit';
COMMENT ON COLUMN users.customer_active_credit_package.customer_id IS 'Customer who purchased the package (PRIMARY KEY ensures one package per customer)';
COMMENT ON COLUMN users.customer_active_credit_package.credit_package_id IS 'The credit package they purchased';
COMMENT ON COLUMN users.customer_active_credit_package.weekly_credit_limit IS 'Weekly credit limit from the package (copied here for performance)';
COMMENT ON COLUMN users.customer_active_credit_package.purchased_at IS 'When this package was purchased';

-- Add index for package lookups
CREATE INDEX IF NOT EXISTS idx_customer_active_credit_package_package_id ON users.customer_active_credit_package(credit_package_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_customer_active_credit_package_package_id;
DROP TABLE IF EXISTS users.customer_active_credit_package;

-- +goose StatementEnd
