-- +goose Up
-- +goose StatementBegin
-- Add soft delete columns to users table
ALTER TABLE users.users
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS scheduled_deletion_at TIMESTAMPTZ;

-- Add index for soft delete queries
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users.users(deleted_at) WHERE deleted_at IS NOT NULL;

-- Add comment explaining the soft delete columns
COMMENT ON COLUMN users.users.deleted_at IS 'Timestamp when account was soft deleted. NULL means account is active. Account data kept for recovery period (30-90 days)';
COMMENT ON COLUMN users.users.scheduled_deletion_at IS 'Timestamp when account is scheduled for permanent deletion. Used for grace period recovery';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS users.idx_users_deleted_at;
ALTER TABLE users.users
DROP COLUMN IF EXISTS deleted_at,
DROP COLUMN IF EXISTS scheduled_deletion_at;
-- +goose StatementEnd
