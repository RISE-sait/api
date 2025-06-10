-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS playground;

CREATE TABLE playground.systems (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE playground.sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    system_id UUID NOT NULL,
    customer_id UUID NOT NULL,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_system_id FOREIGN KEY (system_id)
        REFERENCES playground.systems(id) ON DELETE CASCADE,
    CONSTRAINT fk_customer FOREIGN KEY (customer_id)
        REFERENCES users.users(id) ON DELETE CASCADE,
    CONSTRAINT check_end_time CHECK (end_time > start_time),
    CONSTRAINT unique_schedule EXCLUDE USING GIST (
        system_id WITH =,
        tstzrange(start_time, end_time, '[]') WITH &&
    )
);
-- +goose StatementEnd

-- +goose Down
DROP SCHEMA IF EXISTS playground CASCADE;