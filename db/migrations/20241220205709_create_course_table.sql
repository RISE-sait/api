-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS course;

CREATE TABLE course.courses
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name        VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    capacity    INT         NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS course.courses;

DROP SCHEMA IF EXISTS course;
-- +goose StatementEnd