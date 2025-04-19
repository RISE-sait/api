-- +goose Up
-- +goose StatementBegin
ALTER TABLE events.events
    ADD COLUMN is_date_time_modified BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN recurrence_id         UUID,

    -- constraint of is date time modified can only be true if recurrence id is not null
    ADD CONSTRAINT is_date_time_modified_not_recurring CHECK (
        (recurrence_id IS NOT NULL OR is_date_time_modified = FALSE)
        );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events.events
    DROP CONSTRAINT if exists is_date_time_modified_not_recurring,

    DROP COLUMN if exists is_date_time_modified,
    DROP COLUMN if exists recurrence_id;
-- +goose StatementEnd
