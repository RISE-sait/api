-- +goose Up
-- +goose StatementBegin
ALTER TABLE users.pending_users
ADD COLUMN phone varchar(15);

alter table users.users
    DROP COLUMN has_marketing_email_consent,
    DROP COLUMN has_sms_consent;

alter table users.pending_users
    ADD COLUMN has_marketing_email_consent bool NOT null ,
    ADD COLUMN has_sms_consent bool NOT null ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

alter table users.pending_users
    DROP COLUMN has_marketing_email_consent,
    DROP COLUMN has_sms_consent;

alter table users.users
    ADD COLUMN has_marketing_email_consent bool NOT null ,
    ADD COLUMN has_sms_consent bool NOT null ;

ALTER TABLE users.pending_users
DROP COLUMN phone;
-- +goose StatementEnd
