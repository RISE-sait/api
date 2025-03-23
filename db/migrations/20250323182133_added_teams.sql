-- +goose Up
-- +goose StatementBegin
create table IF NOT EXISTS athletic.teams
(
    id         UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    name       varchar(100) NOT NULL,
    capacity   int          NOT NULL,
    created_at timestamptz  NOT NULL default current_timestamp,
    updated_at timestamptz  NOT NULL default current_timestamp,
    coach_id   uuid,
    CONSTRAINT fk_coach FOREIGN KEY (coach_id) REFERENCES staff.staff (id) ON DELETE SET NULL
);

ALTER TABLE athletic.athletes
    ADD COLUMN if not exists team_id uuid,
    ADD CONSTRAINT fk_team FOREIGN KEY (team_id) REFERENCES athletic.teams (id) ON DELETE SET NULL;

ALTER TABlE events.events
    ADD COLUMN if not exists team_id uuid,
    ADD CONSTRAINT fk_team FOREIGN KEY (team_id) REFERENCES athletic.teams (id) ON DELETE SET NULL;

ALTER TAble public.games
    ADD COLUMN if not exists win_team   uuid NOT NULL default '00000000-0000-0000-0000-000000000000',
    ADD COLUMN if not exists loser_team uuid NOT NULL default '00000000-0000-0000-0000-000000000000',
    ADD COLUMN if not exists win_score  int  NOT NULL default 0,
    ADD COLUMN if not exists lose_score int  NOT NULL default 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Revert column changes in public.games
ALTER TABLE public.games
    DROP COLUMN win_team,
    DROP COLUMN loser_team,
    DROP COLUMN win_score,
    DROP COLUMN lose_score;

-- Revert column changes in public.events (previously schedules)
ALTER TABLE events.events
    DROP CONSTRAINT IF EXISTS fk_team, -- Drop the foreign key constraint
    DROP COLUMN team_id;

-- Revert column changes in users.athletes
ALTER TABLE athletic.athletes
    DROP CONSTRAINT IF EXISTS fk_team, -- Drop the foreign key constraint
    DROP COLUMN team_id;

-- Drop the athletic.teams table
DROP TABLE IF EXISTS athletic.teams;

-- +goose StatementEnd