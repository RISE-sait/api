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
    RENAME COLUMN begin_date_time TO event_start_at;

ALTER TABLE public.events
    RENAME COLUMN end_date_time TO event_end_at;

ALTER TABLE public.events
    ADD CONSTRAINT check_event_times CHECK (event_start_at < event_end_at);

ALTER TABLE public.events
    ADD CONSTRAINT check_session_times CHECK (session_start_time < session_end_time);

ALTER TABLE public.events
    ALTER COLUMN location_id DROP NOT NULL;

ALTER TABLE public.events
    ALTER COLUMN created_at SET NOT NULL;

ALTER TABLE public.events
    ALTER COLUMN updated_at SET NOT NULL;

CREATE OR REPLACE FUNCTION check_event_constraint()
    RETURNS TRIGGER
AS $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM events e
        WHERE e.location_id = NEW.location_id
          AND (
            (NEW.event_start_at <= e.event_end_at AND NEW.event_end_at >= e.event_start_at)
            )
          AND (
            (NEW.session_start_time <= e.session_end_time AND NEW.session_end_time >= e.session_start_time)
            )
        AND e.day = NEW.day
    ) THEN
        RAISE EXCEPTION 'An event at this location overlaps with an existing event. Please choose a different time.';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.events
    DROP COLUMN day,
    DROP COLUMN session_start_time,
    DROP COLUMN session_end_time;

ALTER TABLE public.events
    RENAME COLUMN event_start_at TO begin_date_time;

ALTER TABLE public.events
    RENAME COLUMN event_end_at TO end_date_time;

ALTER TABLE public.events
    DROP CONSTRAINT check_event_times;

ALTER TABLE public.events
    ALTER COLUMN location_id SET NOT NULL;

ALTER TABLE public.events
    ALTER COLUMN created_at DROP NOT NULL;

ALTER TABLE public.events
    ALTER COLUMN updated_at DROP NOT NULL;

DROP TYPE day_enum;
-- +goose StatementEnd
