-- +goose Up
-- +goose StatementBegin

-- Insert default programs. No uniqueness enforced on `type`.
INSERT INTO program.programs (id, name, type, description)
VALUES
    (gen_random_uuid(), 'Practice', 'practice', 'Default program for practices'),
    (gen_random_uuid(), 'Course', 'course', 'Default program for courses'),
    (gen_random_uuid(), 'Other', 'other', 'Default program for other events')
ON CONFLICT (name) DO NOTHING;  -- Optional: avoid dup name conflict

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Delete the inserted default rows
DELETE FROM program.programs
WHERE name IN ('Practice', 'Course', 'Other');

-- +goose StatementEnd
