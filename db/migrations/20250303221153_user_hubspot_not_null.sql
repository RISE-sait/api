-- +goose Up
-- +goose StatementBegin
ALTER TABLE users.users
    ALTER COLUMN hubspot_id SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users.users
    ALTER COLUMN hubspot_id DROP NOT NULL;
-- +goose StatementEnd
