-- +goose Up
CREATE TABLE staff_activity_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    activity VARCHAR(1000) NOT NULL, -- Description of the activity
    occurred_at TIMESTAMPTZ DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users.users (id) ON DELETE CASCADE
);


-- +goose Down
DROP TABLE IF EXISTS staff_activity_logs;