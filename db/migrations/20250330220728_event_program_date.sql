-- +goose Up
-- +goose StatementBegin
ALTER TABLE events.events
    RENAME COLUMN program_start_at TO recurrence_start_at;

ALTER TABLE events.events
    RENAME COLUMN program_end_at TO recurrence_end_at;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events.events
    RENAME COLUMN recurrence_start_at TO program_start_at;

ALTER TABLE events.events
    RENAME COLUMN recurrence_end_at TO program_end_at;
-- +goose StatementEnd
