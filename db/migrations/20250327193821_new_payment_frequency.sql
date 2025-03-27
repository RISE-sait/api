-- +goose Up
-- +goose StatementBegin
-- 1. First create a temporary enum with the new values
CREATE TYPE payment_frequency_temp AS ENUM ('once', 'day', 'week', 'biweekly', 'month');

-- 2. Convert the column to use the temporary enum
ALTER TABLE membership.membership_plans
    ALTER COLUMN payment_frequency TYPE payment_frequency_temp
        USING payment_frequency::text::payment_frequency_temp;

-- 3. Drop the old enum
DROP TYPE payment_frequency;

-- 4. Recreate the original enum name with new values
CREATE TYPE payment_frequency AS ENUM ('once', 'day', 'week', 'biweekly', 'month');

-- 5. Convert back to the properly named enum
ALTER TABLE membership.membership_plans
    ALTER COLUMN payment_frequency TYPE payment_frequency
        USING payment_frequency::text::payment_frequency;

-- 6. Drop the temporary enum
DROP TYPE payment_frequency_temp;

-- 7. Update existing 'once' payments
UPDATE membership.membership_plans
SET amt_periods = NULL
WHERE payment_frequency = 'once';

-- 8. Add constraint
ALTER TABLE membership.membership_plans
    ADD CONSTRAINT check_periods_for_once_payments
        CHECK (
            (payment_frequency = 'once' AND amt_periods IS NULL) OR
            (payment_frequency != 'once')
            );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- 1. Remove constraint first
ALTER TABLE membership.membership_plans
    DROP CONSTRAINT IF EXISTS check_periods_for_once_payments;

-- 2. Revert to original enum values
CREATE TYPE payment_frequency_old AS ENUM ('once', 'day', 'week', 'month');

-- 3. Convert column back to original enum values
ALTER TABLE membership.membership_plans
    ALTER COLUMN payment_frequency TYPE payment_frequency_old
        USING (
        CASE payment_frequency::text
            WHEN 'biweekly' THEN 'week'::payment_frequency_old
            ELSE payment_frequency::text::payment_frequency_old
            END
        );

-- 4. Drop the current enum
DROP TYPE payment_frequency;

-- 5. Recreate original enum
CREATE TYPE payment_frequency AS ENUM ('once', 'day', 'week', 'month');

-- 6. Final conversion
ALTER TABLE membership.membership_plans
    ALTER COLUMN payment_frequency TYPE payment_frequency
        USING payment_frequency::text::payment_frequency;

-- 7. Clean up
DROP TYPE payment_frequency_old;
-- +goose StatementEnd