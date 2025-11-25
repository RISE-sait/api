-- +goose Up
-- +goose StatementBegin

-- Drop the level column from programs table
ALTER TABLE program.programs DROP COLUMN IF EXISTS level;

-- Drop the program_level enum type
DROP TYPE IF EXISTS program.program_level;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Recreate the program_level enum type
CREATE TYPE program.program_level AS ENUM ('beginner', 'intermediate', 'advanced', 'all');

-- Add the level column back with default value
ALTER TABLE program.programs ADD COLUMN level program.program_level NOT NULL DEFAULT 'all';

-- +goose StatementEnd
