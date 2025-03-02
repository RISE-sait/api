-- +goose Up
-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS location;

CREATE TABLE location.facility_categories
(
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL
);

CREATE TABLE location.facilities
(
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                 VARCHAR(255) UNIQUE NOT NULL,
    address              VARCHAR(255)        NOT NULL,
    facility_category_id UUID                NOT NULL,
    FOREIGN KEY (facility_category_id) REFERENCES location.facility_categories (id) ON DELETE CASCADE
);

CREATE TABLE location.locations
(
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                 VARCHAR(100) UNIQUE NOT NULL,
    facility_id UUID                NOT NULL,
    FOREIGN KEY (facility_id) REFERENCES location.facilities (id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS location.locations;

DROP TABLE IF EXISTS location.facilities;

DROP TABLE IF EXISTS location.facility_categories;

DROP SCHEMA IF EXISTS location;
-- +goose StatementEnd
