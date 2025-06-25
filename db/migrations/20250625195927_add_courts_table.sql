-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS location.courts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    location_id UUID NOT NULL REFERENCES location.locations(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_location_court_name UNIQUE(location_id, name)
);

ALTER TABLE events.events ADD COLUMN IF NOT EXISTS court_id UUID REFERENCES location.courts(id);
ALTER TABLE game.games ADD COLUMN IF NOT EXISTS court_id UUID REFERENCES location.courts(id);

ALTER TABLE events.events DROP CONSTRAINT IF EXISTS no_overlapping_events;
ALTER TABLE events.events
    ADD CONSTRAINT no_overlapping_events
        EXCLUDE USING GIST (
            court_id WITH =,
            tstzrange(start_at, end_at) WITH &&
        );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events.events DROP CONSTRAINT IF EXISTS no_overlapping_events;
ALTER TABLE events.events
    ADD CONSTRAINT no_overlapping_events
        EXCLUDE USING GIST (
            location_id WITH =,
            tstzrange(start_at, end_at) WITH &&
        );
ALTER TABLE game.games DROP COLUMN IF EXISTS court_id;
ALTER TABLE events.events DROP COLUMN IF EXISTS court_id;
DROP TABLE IF EXISTS location.courts;
-- +goose StatementEnd