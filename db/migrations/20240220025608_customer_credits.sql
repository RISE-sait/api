-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users.customer_credits
(
    customer_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    credits     INT NOT NULL     DEFAULT 0,
    CONSTRAINT fk_customer FOREIGN KEY (customer_id) REFERENCES users.users (id) ON DELETE CASCADE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users.customer_credits;
-- +goose StatementEnd