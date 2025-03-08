-- +goose Up
-- +goose StatementBegin
ALTER TABLE users.users
    ADD COLUMN has_marketing_email_consent bool NOT NULL,
    ADD COLUMN has_sms_consent             bool NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users.users
    DROP COLUMN has_marketing_email_consent,
    DROP COLUMN has_sms_consent;
-- +goose StatementEnd