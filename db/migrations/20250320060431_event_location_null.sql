-- +goose Up
-- +goose StatementBegin
ALTER TABLE public.events
    ALTER COLUMN location_id DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE public.events
    ALTER COLUMN location_id SET NOT NULL;
-- +goose StatementEnd
