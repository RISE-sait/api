-- +goose Up
-- +goose StatementBegin
ALTER TABLE users.pending_users
    ADD COLUMN alpha_2_country_code CHAR(2);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE users.pending_users
    DROP COLUMN alpha_2_country_code;
-- +goose StatementEnd
