-- +goose Up
-- +goose StatementBegin
-- Add email change columns to users table
ALTER TABLE users.users
ADD COLUMN IF NOT EXISTS pending_email VARCHAR(255),
ADD COLUMN IF NOT EXISTS pending_email_token VARCHAR(255),
ADD COLUMN IF NOT EXISTS pending_email_token_expires_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS email_changed_at TIMESTAMPTZ;

-- Add index for pending email token lookups
CREATE INDEX IF NOT EXISTS idx_users_pending_email_token ON users.users(pending_email_token) WHERE pending_email_token IS NOT NULL;

-- Add comments explaining the email change columns
COMMENT ON COLUMN users.users.pending_email IS 'The new email address awaiting verification before it replaces the current email.';
COMMENT ON COLUMN users.users.pending_email_token IS 'One-time token sent to new email for verification. NULL after change is complete.';
COMMENT ON COLUMN users.users.pending_email_token_expires_at IS 'Expiration time for email change token. Tokens are valid for 24 hours.';
COMMENT ON COLUMN users.users.email_changed_at IS 'Timestamp when the user last changed their email address.';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS users.idx_users_pending_email_token;
ALTER TABLE users.users
DROP COLUMN IF EXISTS pending_email,
DROP COLUMN IF EXISTS pending_email_token,
DROP COLUMN IF EXISTS pending_email_token_expires_at,
DROP COLUMN IF EXISTS email_changed_at;
-- +goose StatementEnd
