-- +goose Up
CREATE TABLE staff_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_name TEXT NOT NULL UNIQUE
);

CREATE Table users.staff (
    id UUID PRIMARY KEY REFERENCES users.users(id),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    role_id UUID NOT NULL REFERENCES staff_roles(id)
);

-- +goose Down
DROP TABLE IF EXISTS users.staff;

DROP TABLE IF EXISTS staff_roles;