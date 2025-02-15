-- +goose Up
CREATE TABLE customers (
    user_id UUID PRIMARY KEY,
    hubspot_id BIGINT UNIQUE NOT NULL,
    credits INT NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);


-- +goose Down
DROP TABLE IF EXISTS customers;
