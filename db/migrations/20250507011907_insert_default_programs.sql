-- +goose Up
-- +goose StatementBegin

-- Safely insert defaults without enforcing uniqueness on type
INSERT INTO program.programs (id, name, type, description)
VALUES
    (gen_random_uuid(), 'Game', 'game', 'Default program for games'),
    (gen_random_uuid(), 'Practice', 'practice', 'Default program for practices'),
    (gen_random_uuid(), 'Course', 'course', 'Default program for courses'),
    (gen_random_uuid(), 'Other', 'other', 'Default program for other events')
ON CONFLICT (type) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM program.programs
WHERE type IN ('game', 'practice', 'course', 'other');
-- +goose StatementEnd
