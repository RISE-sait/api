-- +goose Up
-- +goose StatementBegin

-- Add is_visible column to membership_plans to allow hiding plans without deleting them
ALTER TABLE membership.membership_plans
ADD COLUMN is_visible BOOLEAN NOT NULL DEFAULT true;

-- Add index for better query performance when filtering by visibility
CREATE INDEX idx_membership_plans_is_visible ON membership.membership_plans(is_visible);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove the index and column
DROP INDEX IF EXISTS membership.idx_membership_plans_is_visible;
ALTER TABLE membership.membership_plans
DROP COLUMN IF EXISTS is_visible;

-- +goose StatementEnd
