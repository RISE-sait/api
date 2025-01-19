-- +goose Up
ALTER TABLE schedules
ADD CONSTRAINT chk_end_after_begin CHECK (end_datetime > begin_datetime);

-- +goose Down
ALTER TABLE schedules
DROP CONSTRAINT chk_end_after_begin;