-- +goose Up
CREATE TABLE facility.facilities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    address VARCHAR(255) NOT NULL,
    facility_category_id UUID NOT NULL,
    FOREIGN KEY (facility_category_id) REFERENCES facility.facility_categories (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS facility.facilities;