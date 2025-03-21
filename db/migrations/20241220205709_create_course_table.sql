-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS course;

CREATE TABLE IF NOT EXISTS course.courses
(
    id          UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    name        VARCHAR(50) NOT NULL UNIQUE,
    description TEXT NOT NULL,
    capacity    INT         NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS course.courses;

DROP SCHEMA IF EXISTS course;
-- +goose StatementEnd