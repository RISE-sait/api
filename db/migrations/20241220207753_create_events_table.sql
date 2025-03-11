-- +goose Up

-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TYPE day_enum AS ENUM ('MONDAY', 'TUESDAY', 'WEDNESDAY', 'THURSDAY', 'FRIDAY', 'SATURDAY', 'SUNDAY');

CREATE TYPE timetzrange AS RANGE
(
    subtype = timetz
);

CREATE TABLE events
(
    id                 UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    event_start_at     TIMESTAMP WITH TIME ZONE NOT NULL,
    event_end_at       TIMESTAMP WITH TIME ZONE NOT NULL,
    session_start_time TIMETZ                   NOT NULL,
    session_end_time   TIMETZ                   NOT NULL,
    day                day_enum                 NOT NULL,
    practice_id        UUID,
    course_id          UUID,
    game_id            UUID,
    location_id        UUID                     NOT NULL,
    created_at         TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ              NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_game FOREIGN KEY (game_id) REFERENCES games (id) ON DELETE cascade,
    CONSTRAINT fk_course FOREIGN KEY (course_id) REFERENCES course.courses (id) ON DELETE cascade,
    CONSTRAINT fk_practice FOREIGN KEY (practice_id) REFERENCES practices (id) ON DELETE cascade,
    CONSTRAINT fk_location FOREIGN KEY (location_id) REFERENCES location.locations (id) ON DELETE cascade,
    CONSTRAINT event_end_after_start CHECK (event_end_at > event_start_at),
    CONSTRAINT session_end_after_start CHECK (session_end_time > session_start_time)
);

-- Index for faster event conflict checks
CREATE INDEX idx_events_location_day_time ON events (location_id, day, event_start_at, event_end_at, session_start_time,
                                                     session_end_time);


ALTER TABLE events
    ADD CONSTRAINT unique_event_time
        EXCLUDE USING GIST (
        location_id WITH =,
        day WITH =,
        tstzrange(event_start_at, event_end_at, '[]') WITH &&,
        timetzrange(session_start_time, session_end_time, '[]') WITH && -- Use the custom timetzrange
        );

-- +goose StatementEnd

-- +goose Down
ALTER TABLE events
    DROP CONSTRAINT IF EXISTS unique_event_time;

DROP INDEX IF EXISTS idx_events_location_day_time;

-- Drop the events table
DROP TABLE IF EXISTS events;

DROP TYPE IF EXISTS day_enum;
