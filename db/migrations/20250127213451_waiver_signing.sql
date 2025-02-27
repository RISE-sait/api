-- +goose Up
CREATE TABLE waiver.waiver_signing (
    user_id UUID NOT NULL,
    waiver_id UUID NOT NULL,
    is_signed BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, waiver_id),
    FOREIGN KEY (user_id) REFERENCES users.users (id) ON DELETE CASCADE,
    FOREIGN KEY (waiver_id) REFERENCES waiver.waiver (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS waiver.waiver_signing;