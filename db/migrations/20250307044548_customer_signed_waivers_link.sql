-- +goose Up
-- +goose StatementBegin
ALTER TABLE waiver.waiver
    ADD COLUMN waiver_name VARCHAR(30) NOT NULL;  -- Now has a length limit

ALTER TABLE waiver.waiver_signing
    ADD COLUMN signed_waiver_link TEXT NOT NULL;

ALTER TABLE waiver.pending_users_waiver_signing
    ADD COLUMN pdf_data BYTEA NOT NULL;

ALTER TABLE waiver.pending_users_waiver_signing
    DROP COLUMN is_signed;

ALTER TABLE waiver.waiver_signing
    DROP COLUMN is_signed;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Reverting the changes made in the `waiver` table
ALTER TABLE waiver.waiver
    DROP COLUMN waiver_name;  -- Dropping the waiver_name column

-- Reverting the changes made in the `waiver_signing` table
ALTER TABLE waiver.waiver_signing
    DROP COLUMN signed_waiver_link;  -- Dropping the signed_waiver_link column

-- Reverting the changes made in the `pending_users_waiver_signing` table
ALTER TABLE waiver.pending_users_waiver_signing
    DROP COLUMN pdf_data;  -- Dropping the pdf_data column

-- Reverting the removal of `is_signed` column
ALTER TABLE waiver.pending_users_waiver_signing
    ADD COLUMN is_signed BOOLEAN NOT NULL;  -- Adding back the is_signed column (ensure it is set to the correct default value if necessary)

ALTER TABLE waiver.waiver_signing
    ADD COLUMN is_signed BOOLEAN NOT NULL;

-- +goose StatementEnd