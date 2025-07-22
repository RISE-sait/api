-- +goose Up
-- +goose NO TRANSACTION
ALTER TYPE program.program_type ADD VALUE IF NOT EXISTS 'tournament';
ALTER TYPE program.program_type ADD VALUE IF NOT EXISTS 'tryouts';
ALTER TYPE program.program_type ADD VALUE IF NOT EXISTS 'event';

-- +goose Down
-- +goose NO TRANSACTION
-- no-op