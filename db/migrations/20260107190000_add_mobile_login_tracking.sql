-- +goose Up
-- +goose StatementBegin

-- Add last_mobile_login_at column to track when users last logged into the mobile app
ALTER TABLE users.users
ADD COLUMN last_mobile_login_at TIMESTAMPTZ NULL;

-- Create index for efficient querying of active mobile users
CREATE INDEX idx_users_last_mobile_login ON users.users(last_mobile_login_at)
WHERE last_mobile_login_at IS NOT NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS users.idx_users_last_mobile_login;
ALTER TABLE users.users DROP COLUMN IF EXISTS last_mobile_login_at;

-- +goose StatementEnd
