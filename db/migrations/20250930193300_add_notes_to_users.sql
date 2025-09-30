-- +goose Up
-- +goose StatementBegin
ALTER TABLE users.users ADD COLUMN notes TEXT;

-- Add index for searching notes
CREATE INDEX idx_users_notes ON users.users USING gin(to_tsvector('english', notes)) WHERE notes IS NOT NULL;

-- Add comment for documentation
COMMENT ON COLUMN users.users.notes IS 'Staff notes about the customer for internal reference';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_notes;
ALTER TABLE users.users DROP COLUMN IF EXISTS notes;
-- +goose StatementEnd