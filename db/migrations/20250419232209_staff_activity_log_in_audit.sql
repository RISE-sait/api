-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS staff.staff_activity_logs CASCADE;

CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE audit.staff_activity_logs
(
    id                   UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    staff_id             uuid        NOT NULL,
    activity_description TEXT        NOT NULL,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    FOREIGN KEY (staff_id) REFERENCES staff.staff (id) ON DELETE CASCADE
);

CREATE INDEX idx_staff_activity_logs_staff_id ON audit.staff_activity_logs (staff_id);
CREATE INDEX idx_staff_activity_logs_created_at ON audit.staff_activity_logs (created_at);
CREATE INDEX idx_staff_activity_logs_activity_description ON audit.staff_activity_logs USING GIN (activity_description gin_trgm_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_staff_activity_logs_staff_id;
DROP INDEX IF EXISTS idx_staff_activity_logs_created_at;
DROP INDEX IF EXISTS idx_staff_activity_logs_activity_description;
DROP TABLE IF EXISTS audit.staff_activity_logs CASCADE;
DROP EXTENSION IF EXISTS pg_trgm;
-- +goose StatementEnd
