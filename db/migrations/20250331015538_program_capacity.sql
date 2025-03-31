-- +goose Up
-- +goose StatementBegin
ALTER TABLE program.programs
    ADD COLUMN if not exists capacity int;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE program.programs
    DROP COLUMN if exists capacity;
-- +goose StatementEnd
