-- +goose Up
ALTER TYPE membership.membership_status ADD VALUE IF NOT EXISTS 'paused';

-- +goose Down
-- Note: PostgreSQL does not support removing enum values directly.
-- To reverse this, you would need to recreate the enum type without 'paused'.
