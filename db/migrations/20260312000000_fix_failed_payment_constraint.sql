-- +goose Up
-- +goose StatementBegin

-- Relax the payment calculation constraint to allow failed payments to store
-- the actual attempted amount. For failed payments, customer_paid is 0 but
-- original_amount should reflect what was attempted.
-- Old constraint: customer_paid = original_amount - discount_amount - subsidy_amount
-- New constraint: same rule, but only enforced for non-failed payments.
ALTER TABLE payments.payment_transactions DROP CONSTRAINT valid_payment_calculation;

ALTER TABLE payments.payment_transactions ADD CONSTRAINT valid_payment_calculation CHECK (
    payment_status = 'failed' OR customer_paid = original_amount - discount_amount - subsidy_amount
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE payments.payment_transactions DROP CONSTRAINT valid_payment_calculation;

ALTER TABLE payments.payment_transactions ADD CONSTRAINT valid_payment_calculation CHECK (
    customer_paid = original_amount - discount_amount - subsidy_amount
);

-- +goose StatementEnd
