-- +goose Up
-- +goose StatementBegin
CREATE table discounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    description TEXT,
    discount_percent INT NOT NULL,
    is_use_unlimited BOOLEAN NOT NULL DEFAULT FALSE,
    use_per_client INT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    valid_from TIMESTAMP WITH TIME ZONE NOT NULL,
    valid_to TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT check_use_per_client CHECK (
        (is_use_unlimited = true) OR 
        (is_use_unlimited = false AND use_per_client IS NOT NULL)
    )
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS discounts;
-- +goose StatementEnd
