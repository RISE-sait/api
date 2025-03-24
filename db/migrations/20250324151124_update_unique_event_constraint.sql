-- +goose Up
-- +goose StatementBegin
-- Drop the existing constraint
ALTER TABLE events.events
    DROP CONSTRAINT unique_event_time;

CREATE TYPE timetzrange AS RANGE
(
    subtype = timetz
);

-- Add the updated constraint
ALTER TABLE events.events
    ADD CONSTRAINT unique_event_time
        EXCLUDE USING GIST (
        location_id WITH =,
        tstzrange(program_start_at, program_end_at, '[]') WITH &&,
        day WITH =,
        timetzrange(event_start_time, event_end_time, '[]') WITH &&
        );

-- Drop the trigger
DROP TRIGGER IF EXISTS check_event_overlap ON events.events;

-- Drop the function
DROP FUNCTION IF EXISTS events.check_event_constraint;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Recreate the function
CREATE OR REPLACE FUNCTION events.check_event_constraint()
    RETURNS TRIGGER
AS
$$
BEGIN
    IF EXISTS (SELECT 1
               FROM events.events e
               WHERE e.location_id = NEW.location_id
                 AND (
                   (NEW.program_start_at <= e.program_end_at AND NEW.program_end_at >= e.program_start_at)
                   )
                 AND (
                   (NEW.event_start_time <= e.event_end_time AND NEW.event_end_time >= e.event_start_time)
                   )
                 AND e.day = NEW.day) THEN
        RAISE EXCEPTION 'An event at this location overlaps with an existing event. Please choose a different time.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Recreate the trigger
CREATE TRIGGER check_event_overlap
    BEFORE INSERT OR UPDATE
    ON events.events
    FOR EACH ROW
EXECUTE FUNCTION events.check_event_constraint();

-- Drop the updated constraint
ALTER TABLE events.events
    DROP CONSTRAINT unique_event_time;

-- Recreate the original constraint
ALTER TABLE events.events
    ADD CONSTRAINT unique_event_time
        EXCLUDE USING GIST (
        location_id WITH =,
        tstzrange(program_start_at, program_end_at, '[]') WITH &&
        );

-- +goose StatementEnd
