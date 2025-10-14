-- +goose Up
-- +goose StatementBegin
-- Add email verification columns to users table
ALTER TABLE users.users
ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS email_verification_token VARCHAR(255),
ADD COLUMN IF NOT EXISTS email_verification_token_expires_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ;

-- Add index for email verification token lookups
CREATE INDEX IF NOT EXISTS idx_users_email_verification_token ON users.users(email_verification_token) WHERE email_verification_token IS NOT NULL;

-- Add index for finding unverified users
CREATE INDEX IF NOT EXISTS idx_users_email_verified ON users.users(email_verified) WHERE email_verified = FALSE;

-- Add comments explaining the email verification columns
COMMENT ON COLUMN users.users.email_verified IS 'Whether the user has verified their email address. Users must verify email before they can log in.';
COMMENT ON COLUMN users.users.email_verification_token IS 'One-time token sent to user email for verification. NULL after verification.';
COMMENT ON COLUMN users.users.email_verification_token_expires_at IS 'Expiration time for verification token. Tokens are valid for 24 hours.';
COMMENT ON COLUMN users.users.email_verified_at IS 'Timestamp when the user verified their email address.';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS users.idx_users_email_verification_token;
DROP INDEX IF EXISTS users.idx_users_email_verified;
ALTER TABLE users.users
DROP COLUMN IF EXISTS email_verified,
DROP COLUMN IF EXISTS email_verification_token,
DROP COLUMN IF EXISTS email_verification_token_expires_at,
DROP COLUMN IF EXISTS email_verified_at;
-- +goose StatementEnd
