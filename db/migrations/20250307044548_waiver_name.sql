-- +goose Up
-- +goose StatementBegin
ALTER TABLE waiver.waiver
    ADD COLUMN waiver_name VARCHAR(30) NOT NULL;  -- Now has a length limit

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Reverting the changes made in the `waiver` table
ALTER TABLE waiver.waiver
    DROP COLUMN waiver_name;  -- Dropping the waiver_name column

-- +goose StatementEnd