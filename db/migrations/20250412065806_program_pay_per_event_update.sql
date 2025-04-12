-- +goose Up
-- +goose StatementBegin
ALTER TABLE program.programs
    ADD COLUMN IF NOT EXISTS pay_per_event BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE program.fees
    DROP COLUMN IF EXISTS pay_per_event;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE program.fees
    ADD COLUMN IF NOT EXISTS pay_per_event BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE program.programs
    DROP COLUMN IF EXISTS pay_per_event;
-- +goose StatementEnd
