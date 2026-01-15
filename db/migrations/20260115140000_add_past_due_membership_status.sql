-- +goose Up
-- +goose StatementBegin
ALTER TYPE membership.membership_status ADD VALUE IF NOT EXISTS 'past_due';
-- +goose StatementEnd

-- +goose Down
-- Note: PostgreSQL does not support removing enum values directly.
-- The 'past_due' value will remain but can be migrated to 'inactive' if needed.
-- +goose StatementBegin
UPDATE users.customer_membership_plans SET status = 'inactive' WHERE status = 'past_due';
-- +goose StatementEnd
