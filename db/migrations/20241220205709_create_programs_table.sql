-- +goose Up
-- +goose StatementBegin

CREATE schema if not exists program;

CREATE TYPE program.program_level AS ENUM ('beginner', 'intermediate', 'advanced', 'all');
CREATE TYPE program.program_type AS ENUM ('practice', 'course', 'game','others');

CREATE TABLE IF NOT EXISTS program.programs
(
    id          UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    name        VARCHAR(150)             NOT NULL UNIQUE,
    description TEXT                     NOT NULL,
    level       program.program_level    NOT NULL DEFAULT 'all',
    type        program.program_type     NOT NULL,
    capacity int,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP SCHEMA IF EXISTS program cascade;
-- +goose StatementEnd