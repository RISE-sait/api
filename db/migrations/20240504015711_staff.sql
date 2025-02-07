-- +goose Up
CREATE TABLE staff_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_name TEXT NOT NULL UNIQUE
);

CREATE Table Staff (
    id UUID PRIMARY KEY REFERENCES users(id),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    role_id UUID NOT NULL REFERENCES staff_roles(id)
);

-- +goose Down
DROP TABLE IF EXISTS Staff;

DROP TABLE IF EXISTS staff_roles;