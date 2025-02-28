-- +goose Up

-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS barber;

CREATE TABLE barber.barber_events
(
    id              UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    begin_date_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date_time   TIMESTAMP WITH TIME ZONE NOT NULL,
    customer_id     UUID                     NOT NULL,
    barber_id       UUID                     NOT NULL,
    created_at      TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_customer FOREIGN KEY (customer_id) REFERENCES users.users (id),
    CONSTRAINT fk_barber FOREIGN KEY (barber_id) REFERENCES users.staff (id),
    CONSTRAINT check_end_time CHECK (end_date_time > begin_date_time) -- Prevent invalid schedules
);

CREATE OR REPLACE FUNCTION barber.check_event_constraint()
    RETURNS TRIGGER
AS
$$
BEGIN
    IF EXISTS (SELECT 1
               FROM barber.barber_events e
               WHERE e.barber_id = NEW.barber_id
                 AND (
                   (NEW.begin_date_time < e.end_date_time AND NEW.end_date_time > e.begin_date_time)
                   ))
    THEN
        RAISE EXCEPTION 'An event for this barber overlaps with an existing event. Please choose a different time.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create the trigger to enforce the constraint
CREATE TRIGGER trg_check_event_constraint
    BEFORE INSERT OR UPDATE
    ON barber.barber_events
    FOR EACH ROW
EXECUTE FUNCTION barber.check_event_constraint();

-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS trg_check_event_constraint ON barber.barber_events;

-- Drop the function
DROP FUNCTION IF EXISTS barber.check_event_constraint;

-- Drop the 'barber_events' table
DROP TABLE IF EXISTS barber.barber_events;

-- Drop the 'barber' schema (if it's empty)
DROP SCHEMA IF EXISTS barber;