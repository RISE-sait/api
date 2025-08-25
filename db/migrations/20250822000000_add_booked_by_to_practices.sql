-- +goose Up
-- +goose StatementBegin
ALTER TABLE practice.practices 
ADD COLUMN booked_by UUID REFERENCES users.users(id);
-- +goose StatementEnd

-- +goose Down  
-- +goose StatementBegin
ALTER TABLE practice.practices 
DROP COLUMN IF EXISTS booked_by;
-- +goose StatementEnd