-- +goose Up
-- +goose StatementBegin
ALTER TABlE public.event_staff
    rename to session_staff;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABlE public.session_staff
    rename to event_staff;
-- +goose StatementEnd
