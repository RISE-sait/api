-- +goose Up
-- +goose StatementBegin
-- Add archived_at timestamp to track when accounts were archived for auto-deletion
ALTER TABLE users.users
ADD COLUMN IF NOT EXISTS archived_at TIMESTAMPTZ;

-- Backfill: set archived_at for already archived accounts (use updated_at as best guess)
UPDATE users.users
SET archived_at = updated_at
WHERE is_archived = TRUE AND archived_at IS NULL;

-- Add index for archived account cleanup queries
CREATE INDEX IF NOT EXISTS idx_users_archived_at ON users.users(archived_at) WHERE archived_at IS NOT NULL;

COMMENT ON COLUMN users.users.archived_at IS 'Timestamp when account was archived. Archived accounts are permanently deleted after 30 days.';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS users.idx_users_archived_at;
ALTER TABLE users.users DROP COLUMN IF EXISTS archived_at;
-- +goose StatementEnd
