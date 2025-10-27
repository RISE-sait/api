-- +goose Up
-- +goose StatementBegin

-- Add suspension fields to users table
ALTER TABLE users.users
ADD COLUMN IF NOT EXISTS suspended_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS suspension_reason TEXT,
ADD COLUMN IF NOT EXISTS suspended_by UUID REFERENCES staff.staff(id) ON DELETE SET NULL,
ADD COLUMN IF NOT EXISTS suspension_expires_at TIMESTAMPTZ;

-- Add indexes for suspension queries
CREATE INDEX IF NOT EXISTS idx_users_suspended_at ON users.users(suspended_at) WHERE suspended_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_suspension_expires_at ON users.users(suspension_expires_at) WHERE suspension_expires_at IS NOT NULL;

-- Add suspension fields to customer_membership_plans table
ALTER TABLE users.customer_membership_plans
ADD COLUMN IF NOT EXISTS suspended_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS suspension_billing_paused BOOLEAN NOT NULL DEFAULT FALSE;

-- Add index for finding suspended memberships
CREATE INDEX IF NOT EXISTS idx_cmp_suspended_at ON users.customer_membership_plans(suspended_at) WHERE suspended_at IS NOT NULL;

-- Add comments explaining the suspension columns
COMMENT ON COLUMN users.users.suspended_at IS 'Timestamp when user was suspended. NULL means user is not suspended.';
COMMENT ON COLUMN users.users.suspension_reason IS 'Admin-provided reason for suspension (e.g., "Violation of community guidelines", "Non-payment")';
COMMENT ON COLUMN users.users.suspended_by IS 'Staff member who suspended the user';
COMMENT ON COLUMN users.users.suspension_expires_at IS 'When suspension automatically expires. Can be set for any duration (1 month, 12 months, etc). NULL means indefinite suspension.';
COMMENT ON COLUMN users.customer_membership_plans.suspended_at IS 'Timestamp when membership billing was suspended';
COMMENT ON COLUMN users.customer_membership_plans.suspension_billing_paused IS 'Whether billing is paused due to suspension. When true, arrears will accrue.';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove indexes
DROP INDEX IF EXISTS users.idx_users_suspended_at;
DROP INDEX IF EXISTS users.idx_users_suspension_expires_at;
DROP INDEX IF EXISTS users.idx_cmp_suspended_at;

-- Remove columns from customer_membership_plans
ALTER TABLE users.customer_membership_plans
DROP COLUMN IF EXISTS suspended_at,
DROP COLUMN IF EXISTS suspension_billing_paused;

-- Remove columns from users
ALTER TABLE users.users
DROP COLUMN IF EXISTS suspended_at,
DROP COLUMN IF EXISTS suspension_reason,
DROP COLUMN IF EXISTS suspended_by,
DROP COLUMN IF EXISTS suspension_expires_at;

-- +goose StatementEnd
