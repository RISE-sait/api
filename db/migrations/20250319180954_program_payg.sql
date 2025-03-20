-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.practices
    ADD COLUMN payg_price numeric(6, 2);

ALTER TABLE course.courses
    ADD COLUMN payg_price numeric(6, 2);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.practices
    DROP COLUMN payg_price;

ALTER TABLE course.courses
    DROP COLUMN payg_price;
-- +goose StatementEnd
