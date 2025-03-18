-- +goose Up
-- +goose StatementBegin
CREATE TYPE day_enum AS ENUM ('MONDAY', 'TUESDAY', 'WEDNESDAY', 'THURSDAY', 'FRIDAY', 'SATURDAY', 'SUNDAY');

ALTER TABLE public.events
    ADD COLUMN day day_enum NOT NULL;

ALTER TABLE public.events
    ADD COLUMN session_start_time TIMETZ NOT NULL;

ALTER TABLE public.events
    ADD COLUMN session_end_time TIMETZ NOT NULL;

ALTER TABLE public.events
    RENAME COLUMN event_start_at TO program_start_at;

ALTER TABLE public.events
    RENAME COLUMN event_end_at TO program_end_at;

ALTER TABLE public.events
    ADD CONSTRAINT check_event_times CHECK (program_start_at < program_end_at);

ALTER TABLE public.events
    ADD CONSTRAINT check_session_times CHECK (session_start_time < session_end_time);

ALTER TABLE events
    DROP CONSTRAINT IF EXISTS unique_event_date_time;

CREATE OR REPLACE FUNCTION check_event_constraint()
    RETURNS TRIGGER
AS
$$
BEGIN
    IF EXISTS (SELECT 1
               FROM events e
               WHERE e.location_id = NEW.location_id
                 AND (
                   (NEW.program_start_at <= e.program_end_at AND NEW.program_end_at >= e.program_start_at)
                   )
                 AND (
                   (NEW.session_start_time <= e.session_end_time AND NEW.session_end_time >= e.session_start_time)
                   )
                 AND e.day = NEW.day) THEN
        RAISE EXCEPTION 'An event at this location overlaps with an existing event. Please choose a different time.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER check_event_overlap
    BEFORE INSERT OR UPDATE
    ON public.events
    FOR EACH ROW
EXECUTE FUNCTION check_event_constraint();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.events
    DROP CONSTRAINT IF EXISTS unique_event_time;
DROP INDEX IF EXISTS events_gist_index;

ALTER TABLE public.events
    DROP COLUMN day,
    DROP COLUMN session_start_time,
    DROP COLUMN session_end_time;

ALTER TABLE public.events
    RENAME COLUMN program_start_at TO event_start_at;

ALTER TABLE public.events
    RENAME COLUMN program_end_at TO event_end_at;

ALTER TABLE public.events
    DROP CONSTRAINT IF EXISTS check_event_times;

ALTER TABLE public.events
    DROP CONSTRAINT IF EXISTS check_session_times;

DROP TYPE IF EXISTS day_enum;

DROP TRIGGER IF EXISTS check_event_overlap ON public.events;
-- +goose StatementEnd