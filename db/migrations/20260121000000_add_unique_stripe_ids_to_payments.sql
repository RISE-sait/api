-- +goose Up
-- +goose StatementBegin

-- Add unique constraints to prevent duplicate payment records
-- Note: PostgreSQL unique constraints allow multiple NULLs, so transactions
-- without these IDs (e.g., one-time payments without invoices) are unaffected

-- Unique constraint on stripe_invoice_id to prevent duplicate invoice records
CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_transactions_unique_stripe_invoice
ON payments.payment_transactions(stripe_invoice_id)
WHERE stripe_invoice_id IS NOT NULL;

-- Unique constraint on stripe_checkout_session_id to prevent duplicate checkout records
CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_transactions_unique_stripe_checkout_session
ON payments.payment_transactions(stripe_checkout_session_id)
WHERE stripe_checkout_session_id IS NOT NULL;

-- Unique constraint on stripe_payment_intent_id to prevent duplicate payment intent records
CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_transactions_unique_stripe_payment_intent
ON payments.payment_transactions(stripe_payment_intent_id)
WHERE stripe_payment_intent_id IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS payments.idx_payment_transactions_unique_stripe_payment_intent;
DROP INDEX IF EXISTS payments.idx_payment_transactions_unique_stripe_checkout_session;
DROP INDEX IF EXISTS payments.idx_payment_transactions_unique_stripe_invoice;

-- +goose StatementEnd
