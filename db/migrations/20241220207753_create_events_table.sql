-- +goose Up

-- +goose StatementBegin

CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE events
(
    id                 UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    event_start_at     TIMESTAMP WITH TIME ZONE NOT NULL,
    event_end_at       TIMESTAMP WITH TIME ZONE NOT NULL,
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
    CONSTRAINT event_end_after_start CHECK (event_end_at > event_start_at)
);

-- Index for faster event conflict checks
CREATE INDEX idx_events_location_time ON events (location_id, event_start_at, event_end_at);


ALTER TABLE events
    ADD CONSTRAINT unique_event_time
        EXCLUDE USING GIST (
        location_id WITH =,
        tstzrange(event_start_at, event_end_at, '[]') WITH &&
        );

-- +goose StatementEnd

-- +goose Down
ALTER TABLE events
    DROP CONSTRAINT IF EXISTS unique_event_time;

DROP INDEX IF EXISTS idx_events_location_time;

-- Drop the events table
DROP TABLE IF EXISTS events;