-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS staff.pending_staff (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    gender CHAR(1) CHECK (gender IN ('M', 'F')) NULL,
    age                         int         NOT NULL,
    phone  varchar(25),
    country_alpha2_code         char(2)     NOT NULL,
    role_id                        uuid NOT NULL,
    created_at TIMESTAMP DEFAULT current_timestamp,
    updated_at TIMESTAMP DEFAULT current_timestamp,

    CONSTRAINT email_unique UNIQUE (email),
    CONSTRAINT phone_unique UNIQUE (phone),
    CONSTRAINT country_alpha2_code_check CHECK (country_alpha2_code ~ '^[A-Z]{2}$'),
    CONSTRAINT fk_role FOREIGN KEY (role_id)
        REFERENCES staff.staff_roles(id)
        ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS staff.pending_staff;
-- +goose StatementEnd
