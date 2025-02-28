-- +goose Up
CREATE TABLE locations
(
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                 VARCHAR(100) UNIQUE NOT NULL,
    facility_id UUID                NOT NULL,
    FOREIGN KEY (facility_id) REFERENCES facility.facilities (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS locations;