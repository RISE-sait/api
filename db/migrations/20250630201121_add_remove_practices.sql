-- +goose Up
-- +goose StatementBegin

-- Step 1: Ensure "other" program exists for reassignment
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM program.programs WHERE type = 'other') THEN
    INSERT INTO program.programs (id, name, type, description)
    VALUES (gen_random_uuid(), 'Other', 'other', 'Default program for uncategorized events');
  END IF;
END
$$;

-- Step 2: Reassign any 'practice' type rows to 'other' and delete the 'practice' program
UPDATE events.events
SET program_id = (
  SELECT id FROM program.programs WHERE type = 'other' LIMIT 1
)
WHERE program_id IN (
  SELECT id FROM program.programs WHERE type::text = 'practice'
);

DELETE FROM program.programs WHERE type::text = 'practice';

-- Step 3: Drop unique constraint on type if exists
ALTER TABLE program.programs
DROP CONSTRAINT IF EXISTS unique_program_type;

-- Force-delete any remaining rows with type 'practice' to avoid enum cast failure
DELETE FROM program.programs WHERE type::text = 'practice';

-- Step 4: Replace enum without 'practice'
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_type WHERE typname = 'program_type') THEN
    ALTER TYPE program.program_type RENAME TO program_type_old;
    CREATE TYPE program.program_type AS ENUM ('course', 'other');
    ALTER TABLE program.programs
    ALTER COLUMN type TYPE program.program_type
    USING type::text::program.program_type;
    DROP TYPE program.program_type_old;
  END IF;
END
$$;

-- Step 5: Drop auto-generated name constraint and add a named one
ALTER TABLE program.programs
DROP CONSTRAINT IF EXISTS programs_name_key;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM pg_constraint
        WHERE conname = 'unique_program_name'
        AND conrelid = 'program.programs'::regclass
    ) THEN
        ALTER TABLE program.programs
        ADD CONSTRAINT unique_program_name UNIQUE (name);
    END IF;
END
$$;

-- Step 6: Create practice schema and practices table
CREATE SCHEMA IF NOT EXISTS practice;

CREATE TABLE IF NOT EXISTS practice.practices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID NOT NULL REFERENCES athletic.teams(id),
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ,
    location_id UUID NOT NULL REFERENCES location.locations(id),
    court_id UUID NOT NULL REFERENCES location.courts(id),
    status TEXT CHECK (status IN ('scheduled', 'completed', 'canceled')) DEFAULT 'scheduled',
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Step 1: Drop the practices table and practice schema
DROP TABLE IF EXISTS practice.practices;
DROP SCHEMA IF EXISTS practice;

-- Step 2: Re-add 'practice' to the enum
ALTER TYPE program.program_type RENAME TO program_type_temp;
CREATE TYPE program.program_type AS ENUM ('course', 'practice', 'other');
ALTER TABLE program.programs
ALTER COLUMN type TYPE program.program_type
USING type::text::program.program_type;
DROP TYPE program.program_type_temp;

-- Step 3: Reinsert default 'Practice' program row
INSERT INTO program.programs (id, name, type, description)
VALUES (gen_random_uuid(), 'Practice', 'practice', 'Default program for practices');

-- Step 4: Restore unique constraint on type
ALTER TABLE program.programs
ADD CONSTRAINT unique_program_type UNIQUE (type);

-- Step 5: Drop the custom name constraint if needed
ALTER TABLE program.programs
DROP CONSTRAINT IF EXISTS unique_program_name;

-- +goose StatementEnd