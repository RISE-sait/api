-- +goose Up
-- +goose StatementBegin
ALTER TABLE location.locations

    ADD COLUMN address VARCHAR(255) NOT NULL;

ALTER TABLE location.locations

    DROP COLUMN facility_id;

DROP TABLE location.facilities;

DROP TABlE location.facility_categories;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
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

ALTER TABLE location.locations
    ADD COLUMN facility_id UUID REFERENCES location.facilities (id) ON DELETE SET NULL;

ALTER TABLE location.locations
    DROP COLUMN address;
-- +goose StatementEnd
