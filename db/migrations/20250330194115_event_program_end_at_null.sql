-- +goose Up
-- +goose StatementBegin
ALTER TABLE events.events
    ALTER COLUMN program_end_at DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events.events
    ALTER COLUMN program_end_at SET NOT NULL;
-- +goose StatementEnd