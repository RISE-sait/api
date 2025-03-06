-- +goose Up
CREATE TABLE users.staff_roles
(
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_name TEXT NOT NULL UNIQUE
);

CREATE Table users.staff
(
    id         UUID PRIMARY KEY REFERENCES users.users (id),
    is_active  BOOLEAN                  NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    role_id    UUID                     NOT NULL REFERENCES users.staff_roles (id)
);

CREATE TABLE users.staff_activity_logs
(
    id          UUID PRIMARY KEY       DEFAULT gen_random_uuid(),
    user_id     UUID          NOT NULL,
    activity    VARCHAR(1000) NOT NULL, -- Description of the activity
    occurred_at TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users.users (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS users.staff_activity_logs;

DROP TABLE IF EXISTS users.staff;

DROP TABLE IF EXISTS users.staff_roles;