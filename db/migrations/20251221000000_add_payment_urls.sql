-- +goose Up
-- +goose StatementBegin

-- Add columns for storing Stripe receipt and invoice URLs
ALTER TABLE payments.payment_transactions
    ADD COLUMN receipt_url TEXT,
    ADD COLUMN invoice_url TEXT,
    ADD COLUMN invoice_pdf_url TEXT;

COMMENT ON COLUMN payments.payment_transactions.receipt_url IS 'Stripe receipt URL for one-time payments (events, programs, credit packages)';
COMMENT ON COLUMN payments.payment_transactions.invoice_url IS 'Stripe hosted invoice URL for subscription payments';
COMMENT ON COLUMN payments.payment_transactions.invoice_pdf_url IS 'Stripe invoice PDF download URL for subscription payments';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE payments.payment_transactions
    DROP COLUMN IF EXISTS receipt_url,
    DROP COLUMN IF EXISTS invoice_url,
    DROP COLUMN IF EXISTS invoice_pdf_url;

-- +goose StatementEnd
