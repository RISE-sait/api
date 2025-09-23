-- +goose Up
-- +goose StatementBegin
ALTER TABLE users.users ADD COLUMN stripe_customer_id VARCHAR(255) NULL;
CREATE INDEX idx_users_stripe_customer_id ON users.users(stripe_customer_id);
ALTER TABLE users.users ADD CONSTRAINT uq_users_stripe_customer_id UNIQUE (stripe_customer_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users.users DROP CONSTRAINT IF EXISTS uq_users_stripe_customer_id;
DROP INDEX IF EXISTS idx_users_stripe_customer_id;
ALTER TABLE users.users DROP COLUMN IF EXISTS stripe_customer_id;
-- +goose StatementEnd