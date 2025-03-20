-- +goose Up

-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE SCHEMA IF NOT EXISTS barber;

CREATE TABLE IF NOT EXISTS barber.barber_events
(
    id              UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    begin_date_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date_time   TIMESTAMP WITH TIME ZONE NOT NULL,
    customer_id     UUID                     NOT NULL,
    barber_id       UUID                     NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_customer FOREIGN KEY (customer_id) REFERENCES users.users (id) ON DELETE cascade,
    CONSTRAINT fk_barber FOREIGN KEY (barber_id) REFERENCES users.staff (id) ON DELETE cascade,
    CONSTRAINT check_end_time CHECK (end_date_time > begin_date_time) -- Prevent invalid schedules
);

ALTER TABLE barber.barber_events
    ADD CONSTRAINT unique_barber_schedule
        EXCLUDE USING GIST (
        barber_id WITH =,
        tstzrange(begin_date_time, end_date_time, '[]') WITH &&
        );

-- +goose StatementEnd

-- +goose Down
ALTER TABLE barber.barber_events
    DROP CONSTRAINT IF EXISTS unique_barber_schedule;

-- Drop the 'barber_events' table
DROP TABLE IF EXISTS barber.barber_events;

-- Drop the 'barber' schema (if it's empty)
DROP SCHEMA IF EXISTS barber;