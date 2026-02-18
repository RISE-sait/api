-- +goose Up
-- +goose StatementBegin
ALTER TABLE users.users ADD COLUMN account_type VARCHAR(20);

-- Backfill children (have a parent_id set)
UPDATE users.users SET account_type = 'child' WHERE parent_id IS NOT NULL;

-- Backfill athletes (exist in athletic.athletes)
UPDATE users.users SET account_type = 'athlete'
WHERE id IN (SELECT id FROM athletic.athletes);

-- Backfill parents (have children linked to them)
UPDATE users.users SET account_type = 'parent'
WHERE id IN (SELECT DISTINCT parent_id FROM users.users WHERE parent_id IS NOT NULL);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users.users DROP COLUMN IF EXISTS account_type;
-- +goose StatementEnd
