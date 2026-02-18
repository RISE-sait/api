-- +goose Up
-- +goose StatementBegin

-- Convert all child account_types to athlete
UPDATE users.users SET account_type = 'athlete' WHERE account_type = 'child';

-- Insert athlete records for users with parent_id who don't have one yet
INSERT INTO athletic.athletes (id)
SELECT u.id FROM users.users u
LEFT JOIN athletic.athletes a ON a.id = u.id
WHERE u.parent_id IS NOT NULL AND a.id IS NULL;

-- Insert athlete records for users with account_type = 'parent' who don't have one yet
INSERT INTO athletic.athletes (id)
SELECT u.id FROM users.users u
LEFT JOIN athletic.athletes a ON a.id = u.id
WHERE u.account_type = 'parent' AND a.id IS NULL;

-- Update existing initiated_by = 'child' rows to 'athlete'
UPDATE users.parent_link_requests SET initiated_by = 'athlete' WHERE initiated_by = 'child';

-- Update CHECK constraint on parent_link_requests.initiated_by
ALTER TABLE users.parent_link_requests DROP CONSTRAINT IF EXISTS parent_link_requests_initiated_by_check;
ALTER TABLE users.parent_link_requests ADD CONSTRAINT parent_link_requests_initiated_by_check
    CHECK (initiated_by IN ('athlete', 'parent'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Revert CHECK constraint
ALTER TABLE users.parent_link_requests DROP CONSTRAINT IF EXISTS parent_link_requests_initiated_by_check;
ALTER TABLE users.parent_link_requests ADD CONSTRAINT parent_link_requests_initiated_by_check
    CHECK (initiated_by IN ('child', 'parent'));

-- Revert initiated_by back to 'child'
UPDATE users.parent_link_requests SET initiated_by = 'child' WHERE initiated_by = 'athlete';

-- Note: We don't remove athlete records or revert account_type in down migration
-- as that could cause data loss for legitimately created athlete records.

-- +goose StatementEnd
