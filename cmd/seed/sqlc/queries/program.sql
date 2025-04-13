-- name: InsertPractices :many
WITH prepared_data as (SELECT unnest(@name_array::text[])                   as name,
                              unnest(@description_array::text[])            as description,
                              unnest(@level_array::program.program_level[]) as level,
                              unnest(@is_pay_per_event_array::boolean[])    AS pay_per_event)
INSERT
INTO program.programs (name, description, type, level, pay_per_event)
SELECT name,
       description,
       'practice',
       level,
       pay_per_event
FROM prepared_data
RETURNING id;

-- name: InsertCourses :exec
WITH prepared_data as (SELECT unnest(@name_array::text[])                   as name,
                              unnest(@description_array::text[])            as description,
                              unnest(@level_array::program.program_level[]) as level,
                              unnest(@is_pay_per_event_array::boolean[])    AS pay_per_event)
INSERT
INTO program.programs (name, description, type, level, pay_per_event)
SELECT name,
       description,
       'course',
       level,
       pay_per_event
FROM prepared_data;

-- name: InsertGames :exec
WITH prepared_data as (SELECT unnest(@name_array::text[])                   as name,
                              unnest(@description_array::text[])            as description,
                              unnest(@level_array::program.program_level[]) as level,
                              unnest(@is_pay_per_event_array::boolean[])    AS pay_per_event),
     game_ids AS (
         INSERT INTO program.programs (name, description, type, level, pay_per_event)
             SELECT name, description, 'game', level, pay_per_event
             FROM prepared_data
             RETURNING id)
INSERT
INTO program.games (id, win_team, lose_team, win_score, lose_score)
VALUES (unnest(ARRAY(SELECT id FROM game_ids)), unnest(@win_team_array::uuid[]), unnest(@lose_team_array::uuid[]),
        unnest(@win_score_array::int[]), unnest(@lose_score_array::int[]));

-- name: InsertProgramFees :exec
WITH prepared_data AS (SELECT unnest(@program_name_array::varchar[])            AS program_name,
                              unnest(@membership_name_array::varchar[])         AS membership_name,
                              unnest(@stripe_program_price_id_array::varchar[]) AS stripe_program_price_id)
INSERT
INTO program.fees (program_id, membership_id, stripe_price_id)
SELECT p.id,
       CASE
           WHEN m.id IS NULL THEN NULL::uuid
           ELSE m.id
           END,
       stripe_program_price_id
FROM prepared_data
         JOIN program.programs p ON p.name = program_name
         LEFT JOIN membership.memberships m ON m.name = membership_name
WHERE stripe_program_price_id <> '';