-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS program_membership
(
    program_id UUID NOT NULL,
    membership_id UUID NOT NULL,
    price_per_booking DECIMAL(4, 2),
    is_eligible BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (program_id, membership_id),
    CONSTRAINT fk_program
        FOREIGN KEY (program_id)
            REFERENCES program.programs (id),
    CONSTRAINT chk_price_if_eligible
        CHECK (
            (is_eligible = TRUE AND price_per_booking IS NOT NULL) OR
            (is_eligible = FALSE)
            )
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS program_membership;

-- +goose StatementEnd
