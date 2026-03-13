-- +goose Up
-- Add CHECK constraint to prevent subsidy balance from going negative
ALTER TABLE subsidies.customer_subsidies
    ADD CONSTRAINT positive_remaining_balance
    CHECK (total_amount_used <= approved_amount);

-- +goose Down
ALTER TABLE subsidies.customer_subsidies
    DROP CONSTRAINT IF EXISTS positive_remaining_balance;
