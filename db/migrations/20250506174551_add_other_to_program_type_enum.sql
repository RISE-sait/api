-- +goose Up
-- +goose StatementBegin
ALTER TYPE program.program_type ADD VALUE IF NOT EXISTS 'other';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'unique_program_type'
    ) THEN
        ALTER TABLE program.programs
        ADD CONSTRAINT unique_program_type UNIQUE (type);
    END IF;
END;
$$ LANGUAGE plpgsql;

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
