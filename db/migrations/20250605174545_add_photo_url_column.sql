-- +goose Up
-- +goose StatementBegin
ALTER TABLE users.customer_membership_plans
ADD COLUMN photo_url TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users.customer_membership_plans
DROP COLUMN IF EXISTS photo_url;
-- +goose StatementEnd