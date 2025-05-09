-- +goose Up
-- +goose StatementBegin

-- Step 1: Ensure 'other' program exists
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM program.programs WHERE type::text = 'other') THEN
    INSERT INTO program.programs (id, name, type, description)
    VALUES (gen_random_uuid(), 'Other', 'other', 'Default program for uncategorized events');
  END IF;
END
$$;

-- Step 2: Safely migrate 'game' program references and delete it only if it exists
DO $$
DECLARE
  other_program_id UUID;
  game_program_id UUID;
BEGIN
  SELECT id INTO other_program_id FROM program.programs WHERE type::text = 'other' LIMIT 1;
  SELECT id INTO game_program_id FROM program.programs WHERE type::text = 'game' LIMIT 1;

  IF game_program_id IS NOT NULL THEN
    UPDATE events.events
    SET program_id = other_program_id
    WHERE program_id = game_program_id;

    DELETE FROM program.programs WHERE id = game_program_id;
  END IF;
END
$$;

-- Step 3: Drop the UNIQUE constraint on type if it exists
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

-- Step 5: Replace auto-generated unique name constraint with a named one
ALTER TABLE program.programs
DROP CONSTRAINT IF EXISTS programs_name_key;

ALTER TABLE program.programs
ADD CONSTRAINT unique_program_name UNIQUE (name);

-- Step 6: Create game schema and games table
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

-- Step 1: Drop the games table and game schema
DROP TABLE IF EXISTS game.games;
DROP SCHEMA IF EXISTS game;

-- Step 2: Recreate the original enum with 'game'
ALTER TYPE program.program_type RENAME TO program_type_temp;

CREATE TYPE program.program_type AS ENUM ('course', 'practice', 'game', 'other');

ALTER TABLE program.programs
ALTER COLUMN type TYPE program.program_type
USING type::text::program.program_type;

DROP TYPE program.program_type_temp;

-- Step 3: Reinsert the default 'Game' program row
INSERT INTO program.programs (id, name, type, description)
VALUES (gen_random_uuid(), 'Game', 'game', 'Default program for games');

-- Step 4: Restore the previous unique constraints
ALTER TABLE program.programs
ADD CONSTRAINT unique_program_type UNIQUE (type);

-- Optional: revert back to unnamed auto-generated name constraint if needed
-- DROP NAMED constraint if it exists
ALTER TABLE program.programs
DROP CONSTRAINT IF EXISTS unique_program_name;


-- +goose StatementEnd
