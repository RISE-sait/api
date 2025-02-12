-- +goose Up
CREATE TABLE facilities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    location VARCHAR(255) NOT NULL,
    facility_type_id UUID NOT NULL,
    FOREIGN KEY (facility_type_id) REFERENCES facility_types (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS facilities;