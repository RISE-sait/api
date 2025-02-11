-- +goose Up
CREATE TABLE user_optional_info (
    id UUID PRIMARY KEY REFERENCES users(id),
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    phone VARCHAR(20) CHECK (phone ~ '^\+?[0-9\s\-]{7,20}$'),
    hashed_password VARCHAR(50)
);

-- +goose Down
DROP TABLE IF EXISTS user_optional_info;