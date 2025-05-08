-- -- name: InsertPractices :many
-- WITH prepared_data as (SELECT unnest(@name_array::text[])                   as name,
--                               unnest(@description_array::text[])            as description,
--                               unnest(@level_array::program.program_level[]) as level,
--                               unnest(@is_pay_per_event_array::boolean[])    AS pay_per_event)
-- INSERT
-- INTO program.programs (name, description, type, level, pay_per_event)
-- SELECT name,
--        description,
--        'practice',
--        level,
--        pay_per_event
-- FROM prepared_data
-- RETURNING id;

-- -- name: InsertCourses :exec
-- WITH prepared_data as (SELECT unnest(@name_array::text[])                   as name,
--                               unnest(@description_array::text[])            as description,
--                               unnest(@level_array::program.program_level[]) as level,
--                               unnest(@is_pay_per_event_array::boolean[])    AS pay_per_event)
-- INSERT
-- INTO program.programs (name, description, type, level, pay_per_event)
-- SELECT name,
--        description,
--        'course',
--        level,
--        pay_per_event
-- FROM prepared_data;
-- -- name: InsertProgramFees :exec
-- WITH prepared_data AS (SELECT unnest(@program_name_array::varchar[])            AS program_name,
--                               unnest(@membership_name_array::varchar[])         AS membership_name,
--                               unnest(@stripe_program_price_id_array::varchar[]) AS stripe_program_price_id)
-- INSERT
-- INTO program.fees (program_id, membership_id, stripe_price_id)
-- SELECT p.id,
--        CASE
--            WHEN m.id IS NULL THEN NULL::uuid
--            ELSE m.id
--            END,
--        stripe_program_price_id
-- FROM prepared_data
--          JOIN program.programs p ON p.name = program_name
--          LEFT JOIN membership.memberships m ON m.name = membership_name
-- WHERE stripe_program_price_id <> '';
-- name: InsertBuiltInPrograms :exec
-- name: InsertBuiltInPrograms :exec
INSERT INTO program.programs (id, name, type, description)
VALUES
    (gen_random_uuid(), 'Practice', 'practice', 'Default program for practices'),
    (gen_random_uuid(), 'Course', 'course', 'Default program for courses'),
    (gen_random_uuid(), 'Other', 'other', 'Default program for other events')
ON CONFLICT (type) DO NOTHING;


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
