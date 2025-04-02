-- +goose Up
-- +goose StatementBegin
CREATE TABLE if not exists enrollment_fees
(
    id            UUID        DEFAULT gen_random_uuid() PRIMARY KEY,
    program_id    UUID                                  NOT NULL REFERENCES program.programs (id) ON DELETE CASCADE,
    membership_id UUID REFERENCES membership.memberships (id) ON DELETE CASCADE,
    drop_in_price NUMERIC(6, 2) CHECK (drop_in_price >= 0),
    program_price NUMERIC(6, 2) CHECK (program_price >= 0),
    created_at    TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at    TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,

    CONSTRAINT AT_least_one_price CHECK (
        (drop_in_price IS NOT NULL OR program_price IS NOT NULL)),
    CONSTRAINT no_duplicate_pricing UNIQUE (program_id, membership_id, drop_in_price)

);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS enrollment_fees;
-- +goose StatementEnd
