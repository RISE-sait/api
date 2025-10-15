-- +goose Up
-- +goose StatementBegin

-- Enable the btree_gist extension for exclusion constraints
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- Create an immutable function to handle NULL end_time
CREATE OR REPLACE FUNCTION practice.get_practice_end_time(start_time timestamptz, end_time timestamptz)
RETURNS timestamptz
LANGUAGE sql
IMMUTABLE
AS $$
    SELECT COALESCE(end_time, start_time + interval '2 hours');
$$;

-- Add constraint to prevent overlapping practices on the same court
-- This allows practices at the same time on different courts, but prevents double-booking the same court
ALTER TABLE practice.practices
ADD CONSTRAINT no_overlapping_practices
EXCLUDE USING gist (
    court_id WITH =,
    tstzrange(start_time, practice.get_practice_end_time(start_time, end_time), '[)') WITH &&
)
WHERE (court_id IS NOT NULL);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove the overlapping practices constraint
ALTER TABLE practice.practices
DROP CONSTRAINT IF EXISTS no_overlapping_practices;

-- Drop the helper function
DROP FUNCTION IF EXISTS practice.get_practice_end_time(timestamptz, timestamptz);

-- +goose StatementEnd
