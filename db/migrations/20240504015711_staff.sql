-- +goose Up

CREATE schema if not exists staff;

CREATE TABLE IF NOT EXISTS staff.staff_roles
(
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_name TEXT NOT NULL UNIQUE
);

CREATE Table IF NOT EXISTS staff.staff
(
    id         UUID PRIMARY KEY REFERENCES users.users (id),
    is_active  BOOLEAN                  NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    role_id UUID NOT NULL REFERENCES staff.staff_roles (id)
);

CREATE TABLE IF NOT EXISTS staff.staff_activity_logs
(
    id          UUID PRIMARY KEY       DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
    activity    VARCHAR(1000) NOT NULL, -- Description of the activity
    occurred_at TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES staff.staff (id) ON DELETE set default
);

-- +goose Down
DROP SCHEMA if exists staff;