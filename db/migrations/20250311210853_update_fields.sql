-- +goose Up
-- +goose StatementBegin
ALTER TABLE membership.membership_plans
    ALTER COLUMN name TYPE varchar(150),
    ALTER COLUMN price TYPE DECIMAL(6, 2),
    ALTER COLUMN joining_fee TYPE DECIMAL(6, 2),
    ALTER COLUMN joining_fee SET NOT NULL,
    ALTER COLUMN payment_frequency SET NOT NULL,
    ALTER COLUMN amt_periods SET NOT NULL,
    ALTER COLUMN created_at SET NOT NULL,
    ALTER COLUMN updated_at SET NOT NULL;

ALTER TABLE membership.memberships
    ALTER COLUMN name TYPE varchar(150);

ALTER TABLE practices
    ALTER COLUMN name TYPE varchar(150);

ALTER TABLE users.users
    ADD COLUMN gender CHAR(1) CHECK (gender IN ('M', 'F')) NULL,
    DROP CONSTRAINT IF EXISTS users_phone_key,
    ALTER COLUMN phone TYPE varchar(25),
    ALTER COLUMN email TYPE varchar(255);

ALTER TABLE practices
    DROP COLUMN should_email_booking_notification;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users.users
    DROP COLUMN IF EXISTS gender,
    ALTER COLUMN phone TYPE varchar(15),
    ALTER COLUMN email TYPE varchar(30);

DO
$$
    BEGIN
        IF NOT EXISTS (SELECT 1
                       FROM pg_constraint
                       WHERE conname = 'users_phone_key') THEN
            ALTER TABLE users.users
                ADD CONSTRAINT users_phone_key UNIQUE (phone);
        END IF;
    END
$$;

ALTER TABLE practices
    ALTER COLUMN name TYPE varchar(50);

ALTER TABLE membership.memberships
    ALTER COLUMN name SET DATA TYPE varchar(50);

ALTER TABLE membership.membership_plans
    ALTER COLUMN joining_fee TYPE INT,
    ALTER COLUMN joining_fee DROP NOT NULL,
    ALTER COLUMN name SET DATA TYPE varchar(50),
    ALTER COLUMN price SET DATA TYPE INT,
    ALTER COLUMN payment_frequency DROP NOT NULL,
    ALTER COLUMN amt_periods DROP NOT NULL,
    ALTER COLUMN created_at DROP NOT NULL,
    ALTER COLUMN updated_at DROP NOT NULL;

AlTER TABLE practices
    ADD COLUMN should_email_booking_notification bool default true;
-- +goose StatementEnd
