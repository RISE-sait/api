-- +goose Up
-- +goose StatementBegin
ALTER TABLE program.program_membership
    ADD COLUMN IF NOT EXISTS pay_per_event BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE program.program_membership
    RENAME COLUMN stripe_program_price_id TO stripe_price_id;

ALTER TABLE program.program_membership
    RENAME TO fees;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE program.fees
    RENAME TO program_membership;

ALTER TABLE program.program_membership
    RENAME COLUMN stripe_price_id TO stripe_program_price_id;

ALTER TABLE program.programs
    DROP COLUMN IF EXISTS pay_per_event;
-- +goose StatementEnd
