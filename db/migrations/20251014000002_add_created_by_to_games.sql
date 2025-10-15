-- +goose Up
-- +goose StatementBegin
-- Add created_by column to track who created/scheduled the game
ALTER TABLE game.games
ADD COLUMN IF NOT EXISTS created_by UUID REFERENCES users.users(id);

-- Add index for faster lookups by creator
CREATE INDEX IF NOT EXISTS idx_games_created_by ON game.games(created_by);

-- Backfill existing games with NULL (historical data)
-- New games will have this field populated automatically
COMMENT ON COLUMN game.games.created_by IS 'User (coach/admin) who created/scheduled this game';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS game.idx_games_created_by;
ALTER TABLE game.games DROP COLUMN IF EXISTS created_by;
-- +goose StatementEnd
