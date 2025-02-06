-- +goose Up
ALTER TABLE schedules
ADD CONSTRAINT chk_end_after_begin CHECK (end_time > begin_time);

-- +goose Down
ALTER TABLE schedules
DROP CONSTRAINT chk_end_after_begin;