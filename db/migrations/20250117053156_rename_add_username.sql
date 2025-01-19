-- +goose Up
-- +goose StatementBegin
ALTER TABLE user_passwords RENAME TO user_optional_info;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE user_optional_info RENAME TO user_passwords;
-- +goose StatementEnd
