-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.practices
    DROP COLUMN start_date,
    DROP COLUMN end_date;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.practices
    ADD COLUMN start_date TIMESTAMPTZ NOT NULL,
    ADD COLUMN end_date TIMESTAMPTZ NOT NULL;
-- +goose StatementEnd