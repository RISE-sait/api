-- +goose Up
-- +goose StatementBegin

-- Add barber availability table to define when barbers work
CREATE TABLE IF NOT EXISTS haircut.barber_availability (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    barber_id UUID NOT NULL,
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6), -- 0=Sunday, 6=Saturday
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_barber_availability 
        FOREIGN KEY (barber_id) 
        REFERENCES staff.staff (id) 
        ON DELETE CASCADE,
    CONSTRAINT check_time_order 
        CHECK (end_time > start_time),
    CONSTRAINT unique_barber_day_time 
        UNIQUE (barber_id, day_of_week, start_time, end_time)
);

-- Index for faster lookups
CREATE INDEX idx_barber_availability_barber_day 
    ON haircut.barber_availability (barber_id, day_of_week);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS haircut.barber_availability;

-- +goose StatementEnd