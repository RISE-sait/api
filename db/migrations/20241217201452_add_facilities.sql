-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS facility;

CREATE TABLE facility.facility_categories
(
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL
);

CREATE TABLE facility.facilities
(
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                 VARCHAR(255) UNIQUE NOT NULL,
    address              VARCHAR(255)        NOT NULL,
    facility_category_id UUID                NOT NULL,
    FOREIGN KEY (facility_category_id) REFERENCES facility.facility_categories (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS facility.facilities;

DROP TABLE IF EXISTS facility.facility_categories;

DROP SCHEMA IF EXISTS facility;
-- +goose StatementEnd
