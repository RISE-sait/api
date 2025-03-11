-- +goose Up

-- +goose StatementBegin

CREATE SCHEMA IF NOT EXISTS audit;

CREATE TYPE audit_status AS ENUM ('PENDING', 'COMPLETED', 'FAILED');

CREATE TABLE audit.outbox
(
    id            UUID PRIMARY KEY      DEFAULT gen_random_uuid(),
    sql_statement TEXT         NOT NULL,                   -- SQL statement for admin review
    status        audit_status NOT NULL DEFAULT 'PENDING', -- Use enum for status
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS audit.outbox;
DROP SCHEMA IF EXISTS audit;
