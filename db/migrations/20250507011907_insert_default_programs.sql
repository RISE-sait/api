-- +goose Up
-- +goose StatementBegin

-- Add a unique constraint so ON CONFLICT (type) works
ALTER TABLE program.programs
ADD CONSTRAINT unique_program_type UNIQUE (type);

-- Safely insert defaults with conflict resolution on type
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

-- Delete the inserted default rows
DELETE FROM program.programs
WHERE type IN ('game', 'practice', 'course', 'other');

-- Drop the unique constraint
ALTER TABLE program.programs
DROP CONSTRAINT IF EXISTS unique_program_type;

-- +goose StatementEnd
