-- +goose Up
-- +goose StatementBegin
ALTER TABLE pending_child_accounts 
ADD COLUMN first_name VARCHAR(50), ADD COLUMN last_name VARCHAR(50);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE pending_child_accounts DROP COLUMN first_name, DROP COLUMN last_name;
-- +goose StatementEnd
