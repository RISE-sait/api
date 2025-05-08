-- +goose Up
-- +goose StatementBegin

-- Step 1: Try to migrate 'game' events to 'other', only if 'game' program exists
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM program.programs WHERE type = 'game') THEN
    UPDATE events.events
    SET program_id = (
      SELECT id FROM program.programs WHERE type = 'other' LIMIT 1
    )
    WHERE program_id = (
      SELECT id FROM program.programs WHERE type = 'game' LIMIT 1
    );
  END IF;
END
$$;

-- Step 2: Delete the 'game' program row
DELETE FROM program.programs WHERE type = 'game';

-- Step 3: Drop the UNIQUE constraint if it exists
ALTER TABLE program.programs
DROP CONSTRAINT IF EXISTS unique_program_type;

-- Step 4: Rename and recreate ENUM only if it exists
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'program_type') THEN
    ALTER TYPE program.program_type RENAME TO program_type_old;
    CREATE TYPE program.program_type AS ENUM ('course', 'practice', 'other');
    ALTER TABLE program.programs
    ALTER COLUMN type TYPE program.program_type
    USING type::text::program.program_type;
    DROP TYPE program.program_type_old;
  END IF;
END
$$;

-- Step 5: Recreate the unique constraint
ALTER TABLE program.programs
ADD CONSTRAINT unique_program_type UNIQUE (type);

-- Step 6: Create game schema + games table
CREATE SCHEMA IF NOT EXISTS game;

CREATE TABLE IF NOT EXISTS game.games (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    home_team_id UUID NOT NULL REFERENCES athletic.teams(id),
    away_team_id UUID NOT NULL REFERENCES athletic.teams(id),
    home_score INT,
    away_score INT,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ,
    location_id UUID NOT NULL REFERENCES location.locations(id),
    status TEXT CHECK (status IN ('scheduled', 'completed', 'canceled')) DEFAULT 'scheduled',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS game.games;
DROP SCHEMA IF EXISTS game;

-- Recreate enum with 'game'
ALTER TYPE program.program_type RENAME TO program_type_temp;

CREATE TYPE program.program_type AS ENUM ('course', 'practice', 'game', 'other');

ALTER TABLE program.programs
ALTER COLUMN type TYPE program.program_type
USING type::text::program.program_type;

DROP TYPE program.program_type_temp;

-- Reinsert default game program row
INSERT INTO program.programs (id, name, type, description)
VALUES (gen_random_uuid(), 'Game', 'game', 'Default program for games');

-- Recreate constraint
ALTER TABLE program.programs
ADD CONSTRAINT unique_program_type UNIQUE (type);

-- +goose StatementEnd
