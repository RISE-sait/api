-- +goose Up
-- +goose StatementBegin

CREATE TYPE payment_status AS ENUM ('pending', 'paid', 'failed');

ALTER TABLE program.customer_enrollment
    ADD COLUMN IF NOT EXISTS payment_status payment_status DEFAULT 'pending' NOT NULL;

ALTER TABLE program.customer_enrollment
    ADD COLUMN IF NOT EXISTS payment_expired_at TIMESTAMPTZ DEFAULT NULL;

ALTER TABLE events.customer_enrollment
    ADD COLUMN IF NOT EXISTS payment_status payment_status DEFAULT 'pending' NOT NULL;

ALTER TABLE events.customer_enrollment
    ADD COLUMN IF NOT EXISTS payment_expired_at TIMESTAMPTZ DEFAULT NULL;

UPDATE program.customer_enrollment
SET payment_status = 'paid'
WHERE payment_status IS NULL;

UPDATE events.customer_enrollment
SET payment_status = 'paid'
WHERE payment_status IS NULL;

CREATE INDEX IF NOT EXISTS idx_customer_enrollment_program_status_expired
    ON program.customer_enrollment (program_id, payment_status, payment_status, payment_expired_at);

CREATE INDEX IF NOT EXISTS idx_customer_enrollment_event_status_expired
    ON events.customer_enrollment (event_id, payment_status, payment_status, payment_expired_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_customer_enrollment_program_status_expired;
DROP INDEX IF EXISTS idx_customer_enrollment_event_status_expired;

-- Remove columns from program.customer_enrollment
ALTER TABLE program.customer_enrollment
    DROP COLUMN IF EXISTS payment_status;

ALTER TABLE program.customer_enrollment
    DROP COLUMN IF EXISTS payment_expired_at;

-- Remove columns from events.customer_enrollment
ALTER TABLE events.customer_enrollment
    DROP COLUMN IF EXISTS payment_status;

ALTER TABLE events.customer_enrollment
    DROP COLUMN IF EXISTS payment_expired_at;

-- Drop the payment_status ENUM type
DROP TYPE IF EXISTS payment_status;
-- +goose StatementEnd
