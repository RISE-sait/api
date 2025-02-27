-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS facility;

CREATE TABLE facility.facility_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS facility.facility_categories;

DROP SCHEMA IF EXISTS facility;
-- +goose StatementEnd
