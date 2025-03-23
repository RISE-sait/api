-- +goose Up
-- +goose StatementBegin
ALTER TABLE events.events
    ADD COLUMN capacity int;

ALTER TABLE public.courses
    DROP COLUMN capacity;

ALTER TABLE public.practices
    DROP COLUMN capacity;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE events.events
    DROP COLUMN capacity;

ALTER TABLE public.courses
    ADD COLUMN capacity int;

ALTER TABLE public.practices
    ADD COLUMN capacity int;
-- +goose StatementEnd
