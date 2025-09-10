-- +goose Up
-- +goose StatementBegin

-- Create credit transaction types enum
CREATE TYPE credit_transaction_type AS ENUM ('enrollment', 'refund', 'purchase', 'admin_adjustment');

-- Create credit transactions table for audit trail
CREATE TABLE IF NOT EXISTS users.credit_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    amount INTEGER NOT NULL, -- negative for deductions, positive for additions
    transaction_type credit_transaction_type NOT NULL,
    event_id UUID NULL REFERENCES events.events(id) ON DELETE SET NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Add indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_credit_transactions_customer_id ON users.credit_transactions (customer_id);
CREATE INDEX IF NOT EXISTS idx_credit_transactions_event_id ON users.credit_transactions (event_id) WHERE event_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_credit_transactions_created_at ON users.credit_transactions (created_at);
CREATE INDEX IF NOT EXISTS idx_credit_transactions_type ON users.credit_transactions (transaction_type);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove indexes
DROP INDEX IF EXISTS idx_credit_transactions_customer_id;
DROP INDEX IF EXISTS idx_credit_transactions_event_id;
DROP INDEX IF EXISTS idx_credit_transactions_created_at;
DROP INDEX IF EXISTS idx_credit_transactions_type;

-- Remove table
DROP TABLE IF EXISTS users.credit_transactions;

-- Remove enum type
DROP TYPE IF EXISTS credit_transaction_type;

-- +goose StatementEnd