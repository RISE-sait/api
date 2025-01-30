-- +goose Up
CREATE TABLE pending_child_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_email VARCHAR(255) NOT NULL,
    user_email VARCHAR(255) NOT NULL,
    password VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    UNIQUE(user_email)
);

-- +goose Down

DROP table if exists pending_child_accounts;