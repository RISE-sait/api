-- +goose Up
-- +goose NO TRANSACTION
ALTER TYPE program.program_type ADD VALUE IF NOT EXISTS 'other';

-- +goose Down
-- +goose NO TRANSACTION

