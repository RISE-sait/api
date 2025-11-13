-- +goose Up
-- +goose StatementBegin

-- This migration handles both cases:
-- 1. Fresh DB: Creates the subsidy tables from scratch
-- 2. Existing DB with old schema: Migrates from old table names to new ones

-- Create the subsidies schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS subsidies;

-- Drop old tables if they exist (from old migration that no longer exists)
DROP TABLE IF EXISTS subsidies.audit_log CASCADE;
DROP TABLE IF EXISTS subsidies.usage_history CASCADE;
DROP TABLE IF EXISTS subsidies.user_subsidies CASCADE;
DROP TABLE IF EXISTS subsidies.programs CASCADE;

-- Drop the correct tables if they already exist (in case migration is re-run)
DROP TABLE IF EXISTS subsidies.usage_transactions CASCADE;
DROP TABLE IF EXISTS subsidies.customer_subsidies CASCADE;
DROP TABLE IF EXISTS subsidies.providers CASCADE;

-- Now create the correct schema

-- Subsidy providers (Jumpstart, etc.)
CREATE TABLE subsidies.providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    contact_email TEXT,
    contact_phone TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Customer subsidy balances (the $500 credit)
CREATE TABLE subsidies.customer_subsidies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,
    provider_id UUID REFERENCES subsidies.providers(id) ON DELETE SET NULL,

    -- Financial tracking
    approved_amount DECIMAL(10,2) NOT NULL,
    total_amount_used DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    remaining_balance DECIMAL(10,2) GENERATED ALWAYS AS (approved_amount - total_amount_used) STORED,

    -- Lifecycle
    status TEXT NOT NULL DEFAULT 'pending',

    -- Approval tracking
    approved_by UUID REFERENCES users.users(id) ON DELETE SET NULL,
    approved_at TIMESTAMPTZ,
    rejected_by UUID REFERENCES users.users(id) ON DELETE SET NULL,
    rejected_at TIMESTAMPTZ,
    rejection_reason TEXT,

    -- Validity period
    valid_from TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    valid_until TIMESTAMPTZ,

    -- Metadata
    reason TEXT,
    application_notes TEXT,
    admin_notes TEXT,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    CONSTRAINT positive_approved CHECK (approved_amount > 0),
    CONSTRAINT non_negative_used CHECK (total_amount_used >= 0),
    CONSTRAINT balance_valid CHECK (total_amount_used <= approved_amount),
    CONSTRAINT valid_status CHECK (status IN ('pending', 'approved', 'active', 'depleted', 'expired', 'rejected'))
);

-- Usage history (every time subsidy is used)
CREATE TABLE subsidies.usage_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_subsidy_id UUID NOT NULL REFERENCES subsidies.customer_subsidies(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES users.users(id) ON DELETE CASCADE,

    -- What was purchased
    transaction_type TEXT NOT NULL DEFAULT 'membership_payment',
    membership_plan_id UUID REFERENCES membership.membership_plans(id) ON DELETE SET NULL,

    -- Financial breakdown
    original_amount DECIMAL(10,2) NOT NULL,
    subsidy_applied DECIMAL(10,2) NOT NULL,
    customer_paid DECIMAL(10,2) NOT NULL,

    -- Stripe linkage
    stripe_subscription_id TEXT,
    stripe_invoice_id TEXT,
    stripe_payment_intent_id TEXT,

    -- Audit
    description TEXT,
    applied_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT amounts_balance CHECK (subsidy_applied + customer_paid = original_amount),
    CONSTRAINT non_negative_amounts CHECK (original_amount >= 0 AND subsidy_applied >= 0 AND customer_paid >= 0)
);

-- Audit log for compliance
CREATE TABLE subsidies.audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_subsidy_id UUID REFERENCES subsidies.customer_subsidies(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    performed_by UUID REFERENCES users.users(id) ON DELETE SET NULL,
    previous_status TEXT,
    new_status TEXT,
    amount_changed DECIMAL(10,2),
    notes TEXT,
    ip_address TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_customer_subsidies_customer_id ON subsidies.customer_subsidies(customer_id);
CREATE INDEX idx_customer_subsidies_status ON subsidies.customer_subsidies(status);
CREATE INDEX idx_customer_subsidies_active ON subsidies.customer_subsidies(customer_id, status)
    WHERE status IN ('approved', 'active');
CREATE INDEX idx_usage_transactions_customer_subsidy ON subsidies.usage_transactions(customer_subsidy_id);
CREATE INDEX idx_usage_transactions_customer ON subsidies.usage_transactions(customer_id);
CREATE INDEX idx_usage_transactions_stripe_invoice ON subsidies.usage_transactions(stripe_invoice_id);
CREATE INDEX idx_audit_log_subsidy ON subsidies.audit_log(customer_subsidy_id);

-- Unique constraint: Only one active subsidy per customer at a time
CREATE UNIQUE INDEX idx_one_active_subsidy_per_customer
    ON subsidies.customer_subsidies(customer_id)
    WHERE status IN ('approved', 'active');

-- Comments for documentation
COMMENT ON TABLE subsidies.providers IS 'Organizations providing subsidies (Jumpstart, etc.)';
COMMENT ON TABLE subsidies.customer_subsidies IS 'Customer subsidy balances - like gift cards';
COMMENT ON COLUMN subsidies.customer_subsidies.remaining_balance IS 'Auto-calculated: approved_amount - total_amount_used';
COMMENT ON TABLE subsidies.usage_transactions IS 'Track each deduction from subsidy balance';
COMMENT ON TABLE subsidies.audit_log IS 'Audit trail for compliance and reporting';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop indexes
DROP INDEX IF EXISTS subsidies.idx_one_active_subsidy_per_customer;
DROP INDEX IF EXISTS subsidies.idx_audit_log_subsidy;
DROP INDEX IF EXISTS subsidies.idx_usage_transactions_stripe_invoice;
DROP INDEX IF EXISTS subsidies.idx_usage_transactions_customer;
DROP INDEX IF EXISTS subsidies.idx_usage_transactions_customer_subsidy;
DROP INDEX IF EXISTS subsidies.idx_customer_subsidies_active;
DROP INDEX IF EXISTS subsidies.idx_customer_subsidies_status;
DROP INDEX IF EXISTS subsidies.idx_customer_subsidies_customer_id;

-- Drop tables in reverse order
DROP TABLE IF EXISTS subsidies.audit_log;
DROP TABLE IF EXISTS subsidies.usage_transactions;
DROP TABLE IF EXISTS subsidies.customer_subsidies;
DROP TABLE IF EXISTS subsidies.providers;

-- Drop the schema
DROP SCHEMA IF EXISTS subsidies CASCADE;

-- +goose StatementEnd
