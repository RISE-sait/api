-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS customer_discount_usage
(
    customer_id UUID NOT NULL,
    discount_id UUID NOT NULL,
    usage_count INT NOT NULL DEFAULT 1,
    last_used_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (customer_id, discount_id),
    CONSTRAINT fk_customer
        FOREIGN KEY (customer_id)
            REFERENCES users.users (id),
    CONSTRAINT fk_discount
        FOREIGN KEY (discount_id)
            REFERENCES discounts (id),
    CONSTRAINT check_usage_count_positive 
        CHECK (usage_count > 0)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS customer_discount_usage;
-- +goose StatementEnd