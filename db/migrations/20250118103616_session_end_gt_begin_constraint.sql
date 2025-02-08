-- +goose Up
ALTER TABLE events
ADD CONSTRAINT chk_end_after_begin CHECK (end_time > begin_time);

-- +goose Down
ALTER TABLE events
DROP CONSTRAINT chk_end_after_begin;
