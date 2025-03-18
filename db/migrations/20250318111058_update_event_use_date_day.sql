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

CREATE INDEX events_gist_index
    ON public.events USING GIST (location_id, day, tstzrange(program_start_at, program_end_at, '[)'),
                                 tstzrange(session_start_time, session_end_time, '[)'));

-- Add exclusion constraint to prevent overlapping session and program times
ALTER TABLE public.events
    ADD CONSTRAINT events_no_overlap
        EXCLUDE USING GIST (
        location_id WITH =,
        day WITH =,
        tstzrange(program_start_at, program_end_at, '[)') WITH &&,
        tstzrange(session_start_time, session_end_time, '[)') WITH &&
        );

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.events
    DROP CONSTRAINT IF EXISTS events_no_overlap;
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
-- +goose StatementEnd