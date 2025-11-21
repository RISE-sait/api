-- +goose Up
-- +goose StatementBegin
-- Add stripe_promotion_code_id column to discounts table
ALTER TABLE discounts ADD COLUMN IF NOT EXISTS stripe_promotion_code_id VARCHAR(255);

-- Add unique constraint for promotion code ID
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'discounts_stripe_promotion_code_id_key'
    ) THEN
        ALTER TABLE discounts ADD CONSTRAINT discounts_stripe_promotion_code_id_key UNIQUE (stripe_promotion_code_id);
    END IF;
END
$$;

-- Add index for promotion code ID lookups
CREATE INDEX IF NOT EXISTS idx_discounts_stripe_promotion_code_id ON discounts(stripe_promotion_code_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop index
DROP INDEX IF EXISTS idx_discounts_stripe_promotion_code_id;

-- Drop constraint
ALTER TABLE discounts DROP CONSTRAINT IF EXISTS discounts_stripe_promotion_code_id_key;

-- Drop column
ALTER TABLE discounts DROP COLUMN IF EXISTS stripe_promotion_code_id;
-- +goose StatementEnd
