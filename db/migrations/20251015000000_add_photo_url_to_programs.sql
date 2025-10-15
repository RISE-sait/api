-- +goose Up
-- +goose StatementBegin
ALTER TABLE program.programs
    ADD COLUMN photo_url VARCHAR(500);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE program.programs
    DROP COLUMN photo_url;
-- +goose StatementEnd
