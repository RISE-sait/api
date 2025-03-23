-- +goose Up
-- +goose StatementBegin
CREATE schema if not exists athletic;
CREATE schema if not exists events;

create table IF NOT EXISTS athletic.teams
(
    id         UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    name       varchar(100) NOT NULL,
    capacity   int          NOT NULL,
    created_at timestamptz  NOT NULL default current_timestamp,
    updated_at timestamptz  NOT NULL default current_timestamp
);

ALTER TABLE athletic.teams
    ADD COLUMN if not exists coach_id uuid,
    ADD CONSTRAINT fk_coach FOREIGN KEY (coach_id) REFERENCES users.staff (id) ON DELETE SET NULL;

ALTER TABLE users.athletes
    ADD COLUMN if not exists team_id uuid,
    ADD CONSTRAINT fk_team FOREIGN KEY (team_id) REFERENCES athletic.teams (id) ON DELETE SET NULL;

ALTER TABlE public.events
    ADD COLUMN if not exists team_id uuid,
    ADD CONSTRAINT fk_team FOREIGN KEY (team_id) REFERENCES athletic.teams (id) ON DELETE SET NULL;

ALTER TAble public.games
    ADD COLUMN if not exists win_team   uuid NOT NULL default '00000000-0000-0000-0000-000000000000',
    ADD COLUMN if not exists loser_team uuid NOT NULL default '00000000-0000-0000-0000-000000000000',
    ADD COLUMN if not exists win_score  int  NOT NULL default 0,
    ADD COLUMN if not exists lose_score int  NOT NULL default 0;

ALTER TABlE public.event_staff
    rename to staff;

ALTER TABlE public.staff
    set schema events;

ALTER TABlE public.customer_enrollment
    set schema events;

ALTER TABlE public.events
    ADD COLUMN capacity int;

ALTER TABLE public.practices
    DROP COLUMN capacity;

ALTER TABLE course.courses
    DROP COLUMN capacity;

ALTER TABLE course.courses
    set schema public;

DROP schema course;

ALTER TABlE public.events
    set schema events;

ALTER TABLE users.athletes
    set schema athletic;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE athletic.athletes
    set schema users;

ALTER TABLE events.events
    SET SCHEMA public;

CREATE schema if not exists course;

ALTER TABLE public.courses
    set schema course;

ALTER TABlE public.events
    DROP COLUMN capacity;

ALTER TABLE public.practices
    ADD COLUMN capacity int;

ALTER TABLE course.courses
    ADD COLUMN capacity int;

-- Revert schema changes
ALTER TABLE events.customer_enrollment
    SET SCHEMA public;

ALTER TABLE events.staff
    SET SCHEMA public;

ALTER TABLE public.staff
    RENAME TO event_staff;

-- Revert column changes in public.games
ALTER TABLE public.games
    DROP COLUMN win_team,
    DROP COLUMN loser_team,
    DROP COLUMN win_score,
    DROP COLUMN lose_score;

-- Revert column changes in public.events (previously schedules)
ALTER TABLE public.events
    DROP CONSTRAINT IF EXISTS fk_team, -- Drop the foreign key constraint
    DROP COLUMN team_id;

-- Revert column changes in users.athletes
ALTER TABLE users.athletes
    DROP CONSTRAINT IF EXISTS fk_team, -- Drop the foreign key constraint
    DROP COLUMN team_id;

-- Drop the athletic.teams table
DROP TABLE IF EXISTS athletic.teams;

-- Drop the schemas
DROP SCHEMA IF EXISTS athletic;
DROP SCHEMA IF EXISTS events;

-- +goose StatementEnd