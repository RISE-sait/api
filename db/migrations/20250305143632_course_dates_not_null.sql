-- +goose Up
-- +goose StatementBegin
ALTER TABLE course.courses
    ALTER COLUMN updated_at SET NOT NULL,
    ALTER COLUMN created_at SET NOT NULL;
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
ALTER TABLE users.users
    ALTER COLUMN updated_at DROP NOT NULL,
    ALTER COLUMN created_at DROP NOT NULL;
-- +goose StatementEnd
