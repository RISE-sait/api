-- +goose Up
-- +goose StatementBegin
ALTER TABLE events.events
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID,
    ADD CONSTRAINT fk_created_by
        FOREIGN KEY (created_by)
            REFERENCES users.users (id)
            ON DELETE SET NULL,
    ADD CONSTRAINT fk_updated_by
        FOREIGN KEY (updated_by)
            REFERENCES users.users (id)
            ON DELETE SET NULL;

ALTER TABLE events.events
    ALTER COLUMN location_id SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE events.events
    ALTER COLUMN location_id DROP NOT NULL;

ALTER TABLE events.events
    DROP COLUMN IF EXISTS created_by,
    DROP COLUMN IF EXISTS updated_by,
    DROP CONSTRAINT IF EXISTS fk_created_by,
    DROP CONSTRAINT IF EXISTS fk_updated_by;
-- +goose StatementEnd
