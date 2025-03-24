-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS games
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    win_team   uuid not null,
    lose_team  uuid not null,
    win_score  int  not null default 0,
    lose_score int  not null default 0,
    CONSTRAINT fk_program_id FOREIGN KEY (id) REFERENCES program.programs (id) ON DELETE cascade,
    CONSTRAINT fk_win_team FOREIGN KEY (win_team) REFERENCES athletic.teams (id) ON DELETE set default,
    CONSTRAINT fk_lose_team FOREIGN KEY (lose_team) REFERENCES athletic.teams (id) ON DELETE set default
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS games;
-- +goose StatementEnd
