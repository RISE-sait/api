-- +goose Up
-- +goose StatementBegin
ALTER TABLE practice.practices 
ALTER COLUMN court_id DROP NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE practice.practices 
ALTER COLUMN court_id SET NOT NULL;
-- +goose StatementEnd