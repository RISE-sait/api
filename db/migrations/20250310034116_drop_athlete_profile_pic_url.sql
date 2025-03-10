-- +goose Up
-- +goose StatementBegin
ALTER TABLE users.athletes
    DROP COLUMN profile_pic_url;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE users.athletes
    ADD COLUMN profile_pic_url TEXT;
-- +goose StatementEnd
