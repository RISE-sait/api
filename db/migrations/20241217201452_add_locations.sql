-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS location;

CREATE TABLE IF NOT EXISTS location.locations
(
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                 VARCHAR(100) UNIQUE NOT NULL,
    address VARCHAR(255) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS location.locations;

DROP SCHEMA IF EXISTS location;
-- +goose StatementEnd
