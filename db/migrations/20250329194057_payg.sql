-- +goose Up
-- +goose StatementBegin
ALTER TABLE program.programs
    ADD COLUMN payg_price numeric(6, 2);

ALTER TABLE public.program_membership
    DROP COLUMN is_eligible;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE public.program_membership
    ADD COLUMN is_eligible boolean NOT NULL DEFAULT false;

ALTER TABLE program.programs
    DROP COLUMN payg_price;
-- +goose StatementEnd
