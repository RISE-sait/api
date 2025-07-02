-- +goose Up
-- +goose StatementBegin
ALTER TABLE users.users
ADD COLUMN IF NOT EXISTS is_archived BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users.users
DROP COLUMN IF EXISTS is_archived;
-- +goose StatementEnd