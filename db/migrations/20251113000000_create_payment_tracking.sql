-- +goose Up
-- +goose StatementBegin

-- Create payments schema for centralized payment tracking
CREATE SCHEMA IF NOT EXISTS payments;

-- Centralized payment tracking table
CREATE TABLE payments.payment_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Customer information
    customer_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    customer_email TEXT NOT NULL,
    customer_name TEXT NOT NULL,

    -- Transaction details
    transaction_type TEXT NOT NULL, -- 'membership_subscription', 'membership_renewal', 'program_enrollment', 'event_registration', 'joining_fee', 'credit_package'
    transaction_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Amounts (using DECIMAL for precision)
    original_amount DECIMAL(10, 2) NOT NULL, -- Original price before any discounts/subsidies
    discount_amount DECIMAL(10, 2) NOT NULL DEFAULT 0.00, -- Amount reduced by discount codes
    subsidy_amount DECIMAL(10, 2) NOT NULL DEFAULT 0.00, -- Amount covered by subsidy
    customer_paid DECIMAL(10, 2) NOT NULL, -- Final amount customer actually paid

    -- Related IDs
    membership_plan_id UUID REFERENCES membership.membership_plans(id) ON DELETE SET NULL,
    program_id UUID REFERENCES program.programs(id) ON DELETE SET NULL,
    event_id UUID REFERENCES events.events(id) ON DELETE SET NULL,
    credit_package_id UUID REFERENCES users.credit_packages(id) ON DELETE SET NULL,
    subsidy_id UUID REFERENCES subsidies.customer_subsidies(id) ON DELETE SET NULL,
    discount_code_id UUID REFERENCES discounts(id) ON DELETE SET NULL,

    -- Stripe information
    stripe_customer_id TEXT,
    stripe_subscription_id TEXT,
    stripe_invoice_id TEXT,
    stripe_payment_intent_id TEXT,
    stripe_checkout_session_id TEXT,

    -- Payment status
    payment_status TEXT NOT NULL DEFAULT 'pending', -- 'pending', 'completed', 'failed', 'refunded', 'partially_refunded'
    payment_method TEXT, -- 'card', 'bank_transfer', etc.
    currency TEXT DEFAULT 'USD',

    -- Metadata
    description TEXT,
    metadata JSONB, -- For storing additional flexible data

    -- Refund tracking
    refunded_amount DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    refund_reason TEXT,
    refunded_at TIMESTAMPTZ,

    -- Audit
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT positive_amounts CHECK (
        original_amount >= 0 AND
        discount_amount >= 0 AND
        subsidy_amount >= 0 AND
        customer_paid >= 0 AND
        refunded_amount >= 0
    ),
    CONSTRAINT valid_payment_calculation CHECK (
        customer_paid = original_amount - discount_amount - subsidy_amount
    ),
    CONSTRAINT valid_payment_status CHECK (
        payment_status IN ('pending', 'completed', 'failed', 'refunded', 'partially_refunded')
    )
);

-- Create indexes for performance
CREATE INDEX idx_payment_transactions_customer_id ON payments.payment_transactions(customer_id);
CREATE INDEX idx_payment_transactions_transaction_date ON payments.payment_transactions(transaction_date DESC);
CREATE INDEX idx_payment_transactions_transaction_type ON payments.payment_transactions(transaction_type);
CREATE INDEX idx_payment_transactions_payment_status ON payments.payment_transactions(payment_status);
CREATE INDEX idx_payment_transactions_stripe_invoice_id ON payments.payment_transactions(stripe_invoice_id);
CREATE INDEX idx_payment_transactions_stripe_subscription_id ON payments.payment_transactions(stripe_subscription_id);
CREATE INDEX idx_payment_transactions_subsidy_id ON payments.payment_transactions(subsidy_id);
CREATE INDEX idx_payment_transactions_created_at ON payments.payment_transactions(created_at DESC);

-- Create composite indexes for common report queries
CREATE INDEX idx_payment_transactions_date_status ON payments.payment_transactions(transaction_date DESC, payment_status);
CREATE INDEX idx_payment_transactions_type_date ON payments.payment_transactions(transaction_type, transaction_date DESC);
CREATE INDEX idx_payment_transactions_customer_date ON payments.payment_transactions(customer_id, transaction_date DESC);

-- Update timestamp trigger function
CREATE OR REPLACE FUNCTION payments.update_payment_transactions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Attach trigger to table
CREATE TRIGGER update_payment_transactions_updated_at
    BEFORE UPDATE ON payments.payment_transactions
    FOR EACH ROW
    EXECUTE FUNCTION payments.update_payment_transactions_updated_at();

-- Add comments for documentation
COMMENT ON TABLE payments.payment_transactions IS 'Centralized tracking of all payment transactions including memberships, events, programs, and subsidies';
COMMENT ON COLUMN payments.payment_transactions.original_amount IS 'Original price before any discounts or subsidies';
COMMENT ON COLUMN payments.payment_transactions.discount_amount IS 'Amount reduced by discount codes';
COMMENT ON COLUMN payments.payment_transactions.subsidy_amount IS 'Amount covered by government/organization subsidy';
COMMENT ON COLUMN payments.payment_transactions.customer_paid IS 'Final amount customer actually paid (original - discount - subsidy)';
COMMENT ON COLUMN payments.payment_transactions.metadata IS 'JSON field for storing additional flexible data like plan names, event details, etc.';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop trigger
DROP TRIGGER IF EXISTS update_payment_transactions_updated_at ON payments.payment_transactions;

-- Drop trigger function
DROP FUNCTION IF EXISTS payments.update_payment_transactions_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS payments.idx_payment_transactions_customer_date;
DROP INDEX IF EXISTS payments.idx_payment_transactions_type_date;
DROP INDEX IF EXISTS payments.idx_payment_transactions_date_status;
DROP INDEX IF EXISTS payments.idx_payment_transactions_created_at;
DROP INDEX IF EXISTS payments.idx_payment_transactions_subsidy_id;
DROP INDEX IF EXISTS payments.idx_payment_transactions_stripe_subscription_id;
DROP INDEX IF EXISTS payments.idx_payment_transactions_stripe_invoice_id;
DROP INDEX IF EXISTS payments.idx_payment_transactions_payment_status;
DROP INDEX IF EXISTS payments.idx_payment_transactions_transaction_type;
DROP INDEX IF EXISTS payments.idx_payment_transactions_transaction_date;
DROP INDEX IF EXISTS payments.idx_payment_transactions_customer_id;

-- Drop table
DROP TABLE IF EXISTS payments.payment_transactions;

-- Drop schema
DROP SCHEMA IF EXISTS payments CASCADE;

-- +goose StatementEnd
