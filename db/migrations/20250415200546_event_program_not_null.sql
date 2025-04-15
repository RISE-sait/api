-- +goose Up
-- +goose StatementBegin
ALTER TABLE events.events
    ALTER COLUMN program_id SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events.events
    ALTER COLUMN program_id DROP NOT NULL;
-- +goose StatementEnd
