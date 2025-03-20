-- +goose Up

-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TYPE day_enum AS ENUM ('MONDAY', 'TUESDAY', 'WEDNESDAY', 'THURSDAY', 'FRIDAY', 'SATURDAY', 'SUNDAY');

CREATE TABLE IF NOT EXISTS events
(
    id                 UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    program_start_at TIMESTAMP WITH TIME ZONE NOT NULL,                         -- Renamed from event_start_at
    program_end_at   TIMESTAMP WITH TIME ZONE NOT NULL,                         -- Renamed from event_end_at
    practice_id        UUID,
    course_id          UUID,
    game_id            UUID,
    location_id        UUID                     NOT NULL,
    created_at         TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
    day              day_enum                 NOT NULL,                         -- New column
    event_start_time TIMETZ                   NOT NULL,                         -- New column
    event_end_time   TIMETZ                   NOT NULL,                         -- New column
    CONSTRAINT fk_game FOREIGN KEY (game_id) REFERENCES games (id) ON DELETE cascade,
    CONSTRAINT fk_course FOREIGN KEY (course_id) REFERENCES course.courses (id) ON DELETE cascade,
    CONSTRAINT fk_practice FOREIGN KEY (practice_id) REFERENCES practices (id) ON DELETE cascade,
    CONSTRAINT fk_location FOREIGN KEY (location_id) REFERENCES location.locations (id) ON DELETE cascade,
    CONSTRAINT event_end_after_start CHECK (program_end_at > program_start_at), -- Updated constraint
    CONSTRAINT check_event_times CHECK (event_start_time < event_end_time)      -- New constraint
);

-- Index for faster event conflict checks
CREATE INDEX idx_events_location_time ON events (location_id, program_start_at, program_end_at);


ALTER TABLE events
    ADD CONSTRAINT unique_event_time
        EXCLUDE USING GIST (
        location_id WITH =,
        tstzrange(program_start_at, program_end_at, '[]') WITH &&
        );

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
                   (NEW.event_start_time <= e.event_end_time AND NEW.event_end_time >= e.event_start_time)
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
DROP TRIGGER IF EXISTS check_event_overlap ON public.events;
DROP FUNCTION IF EXISTS check_event_constraint;

-- Drop the unique_event_time constraint
ALTER TABLE events
    DROP CONSTRAINT IF EXISTS unique_event_time;

-- Drop the index
DROP INDEX IF EXISTS idx_events_location_time;

-- Drop the day_enum type
DROP TYPE IF EXISTS day_enum;

-- Drop the events table
DROP TABLE IF EXISTS events;
