-- +goose Up
CREATE TABLE user_passwords (
    id UUID PRIMARY KEY REFERENCES users(id),
    name VARCHAR(50),
    hashed_password VARCHAR(50)
);

-- +goose Down
DROP TABLE IF EXISTS user_passwords;