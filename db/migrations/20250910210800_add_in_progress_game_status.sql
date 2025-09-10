-- +goose Up
-- +goose StatementBegin

-- Drop the existing check constraint
ALTER TABLE game.games DROP CONSTRAINT IF EXISTS games_status_check;

-- Add the new check constraint with 'in_progress' status
ALTER TABLE game.games ADD CONSTRAINT games_status_check 
    CHECK (status = ANY (ARRAY['scheduled'::text, 'in_progress'::text, 'completed'::text, 'canceled'::text]));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Revert to the original constraint without 'in_progress'
ALTER TABLE game.games DROP CONSTRAINT IF EXISTS games_status_check;
ALTER TABLE game.games ADD CONSTRAINT games_status_check 
    CHECK (status = ANY (ARRAY['scheduled'::text, 'completed'::text, 'canceled'::text]));

-- +goose StatementEnd