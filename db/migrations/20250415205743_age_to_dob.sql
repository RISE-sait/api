-- +goose Up
-- +goose StatementBegin
ALTER TABLE users.users
    ADD COLUMN dob DATE NOT NULL DEFAULT '2000-01-01';

ALTER TABLE staff.pending_staff
    ADD COLUMN dob DATE NOT NULL DEFAULT '2000-01-01';

ALTER TABLE users.users
    DROP COLUMN age;

ALTER TABLE staff.pending_staff
    DROP COLUMN age;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE staff.pending_staff
    ADD COLUMN age INT NOT NULL DEFAULT 1;

ALTER TABLE staff.pending_staff
    DROP COLUMN dob;

ALTER TABLE users.users
    ADD COLUMN age INT NOT NULL DEFAULT 1;

ALTER TABLE users.users
    DROP COLUMN dob;
-- +goose StatementEnd