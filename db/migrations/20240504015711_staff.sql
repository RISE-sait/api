-- +goose Up
CREATE TYPE staff_role_enum AS ENUM ('ADMIN', 'INSTRUCTOR', 'SUPERADMIN','COACH');

CREATE Table Staff (
    id UUID PRIMARY KEY REFERENCES users(id),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    role staff_role_enum NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS Staff;

DROP TYPE IF EXISTS staff_role_enum;