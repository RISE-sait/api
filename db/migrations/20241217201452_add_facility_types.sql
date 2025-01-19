-- +goose Up
CREATE TABLE facility_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS facility_types;

