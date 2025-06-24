-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS membership.discount_restricted_membership_plans
(
    discount_id UUID NOT NULL,
    membership_plan_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (discount_id, membership_plan_id),
    CONSTRAINT fk_discount
        FOREIGN KEY (discount_id)
        REFERENCES discounts(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_membership_plan
        FOREIGN KEY (membership_plan_id)
        REFERENCES membership.membership_plans(id)
        ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS membership.discount_restricted_membership_plans;
-- +goose StatementEnd