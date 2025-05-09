-- +goose Up
ALTER TABLE athletic.teams
ADD COLUMN logo_url TEXT;

-- +goose Down
ALTER TABLE athletic.teams
DROP COLUMN logo_url;
