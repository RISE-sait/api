-- +goose Up
-- +goose StatementBegin
ALTER TABLE athletic.teams
    ADD CONSTRAINT unique_team_name UNIQUE (name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE athletic.teams
    DROP CONSTRAINT unique_team_name;
-- +goose StatementEnd
