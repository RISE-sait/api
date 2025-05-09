-- name: InsertBuiltInPrograms :exec
INSERT INTO program.programs (id, name, type, description)
VALUES
    (gen_random_uuid(), 'Practice', 'practice'::program.program_type, 'Default program for practices'),
    (gen_random_uuid(), 'Course', 'course'::program.program_type, 'Default program for courses'),
    (gen_random_uuid(), 'Other', 'other'::program.program_type, 'Default program for other events')
ON CONFLICT (name) DO UPDATE
SET type = EXCLUDED.type,
    description = EXCLUDED.description;


-- name: InsertProgramFees :exec
WITH prepared_data AS (
    SELECT
        unnest(@program_name_array::varchar[])            AS program_name,
        unnest(@membership_name_array::varchar[])         AS membership_name,
        unnest(@stripe_program_price_id_array::varchar[]) AS stripe_program_price_id
)
INSERT INTO program.fees (program_id, membership_id, stripe_price_id)
SELECT
    p.id,
    CASE WHEN m.id IS NULL THEN NULL::uuid ELSE m.id END,
    stripe_program_price_id
FROM prepared_data
JOIN program.programs p ON p.name = program_name
LEFT JOIN membership.memberships m ON m.name = membership_name
WHERE stripe_program_price_id <> '';

-- name: GetProgramByType :one
SELECT *
FROM program.programs
WHERE type = $1
LIMIT 1;
