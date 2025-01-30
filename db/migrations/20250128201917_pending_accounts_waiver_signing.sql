-- +goose Up
CREATE TABLE pending_accounts_waiver_signing (
    user_id UUID NOT NULL,
    waiver_id UUID NOT NULL,
    is_signed BOOLEAN NOT NULL DEFAULT FALSE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, waiver_id),
    FOREIGN KEY (user_id) REFERENCES pending_child_accounts (id) ON DELETE CASCADE,
    FOREIGN KEY (waiver_id) REFERENCES waiver (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE IF EXISTS pending_accounts_waiver_signing;