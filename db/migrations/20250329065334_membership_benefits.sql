-- +goose Up
-- +goose StatementBegin
ALTER TABLE membership.memberships
    ADD COLUMN benefits varchar(750) NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE membership.memberships
    DROP COLUMN benefits;
-- +goose StatementEnd
