-- +goose Up
-- +goose StatementBegin

ALTER TABLE athletic.athletes
ADD COLUMN photo_url TEXT;

ALTER TABLE staff.staff
ADD COLUMN photo_url TEXT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE staff.staff
DROP COLUMN IF EXISTS photo_url;

ALTER TABLE athletic.athletes
DROP COLUMN IF EXISTS photo_url;

-- +goose StatementEnd
