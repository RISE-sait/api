-- +goose Up
-- +goose StatementBegin
ALTER TABLE events.events
DROP COLUMN IF EXISTS capacity;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events.events
ADD COLUMN IF NOT EXISTS capacity INT
-- +goose StatementEnd
