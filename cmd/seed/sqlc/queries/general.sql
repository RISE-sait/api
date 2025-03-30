-- name: InsertLocations :exec
INSERT INTO location.locations (name, address)
VALUES (unnest(@name_array::text[]), unnest(@address_array::text[]))
RETURNING id;

-- name: InsertPractices :many
WITH prepared_data as (
        SELECT unnest(@name_array::text[]) as name,
        unnest(@description_array::text[]) as description,
        unnest(@level_array::program.program_level[]) as level)
INSERT INTO program.programs (name, description, type, level)
SELECT name,
       description,
       'practice',
       level
FROM prepared_data
RETURNING id;

-- name: InsertCourses :exec
WITH prepared_data as (SELECT unnest(@name_array::text[]) as name,
        unnest(@description_array::text[]) as description,
        unnest(@level_array::program.program_level[]) as level)
INSERT INTO program.programs (name, description, type, level)
SELECT name,
       description,
       'course',
       level
FROM prepared_data;

-- name: InsertGames :exec
WITH prepared_data as (
        SELECT unnest(@name_array::text[]) as name,
        unnest(@description_array::text[]) as description,
        unnest(@level_array::program.program_level[]) as level),
game_ids AS (
    INSERT INTO program.programs (name, description, type, level)
    SELECT name, description, 'game', level
    FROM prepared_data
    RETURNING id
)
INSERT INTO public.games (id, win_team, lose_team, win_score, lose_score)
VALUES (unnest(ARRAY(SELECT id FROM game_ids)), unnest(@win_team_array::uuid[]), unnest(@lose_team_array::uuid[]), unnest(@win_score_array::int[]), unnest(@lose_score_array::int[]));

-- name: InsertTeams :many
WITH prepared_data AS (SELECT unnest(@coach_email_array::text[])          AS coach,
                              unnest(@capacity_array::int[])             AS capacity,
                              unnest(@name_array::text[]) AS name)
INSERT
INTO athletic.teams(capacity, coach_id, name)
SELECT capacity, u.id, name
FROM prepared_data
JOIN users.users u ON u.email = coach
RETURNING id;


-- name: InsertWaivers :exec
INSERT INTO waiver.waiver(waiver_url, waiver_name)
VALUES ('https://www.youtube.com/', 'youtube'),
       ('https://www.youtube.com/watch?v=5GTFt8JNwHU', 'video');

-- name: InsertCoachStats :exec
INSERT INTO athletic.coach_stats (coach_id, wins, losses)
VALUES ((SELECT id FROM users.users WHERE email = 'viktor.djurasic+1@abcfitness.com'),
        1,
        1),
       ((SELECT id FROM users.users WHERE email = 'coach@test.com'),
        1,
        2);

-- name: InsertEnrollmentFees :exec
WITH prepared_data AS (SELECT unnest(@program_id_array::uuid[])       AS program_id,
                              unnest(@membership_id_array::uuid[])    AS membership_id,
                              unnest(@drop_in_price_array::numeric[]) AS drop_in_price,
                              unnest(@program_price_array::numeric[]) AS program_price)
INSERT
INTO enrollment_fees (program_id, membership_id, drop_in_price, program_price)
SELECT program_id,
       CASE
           WHEN membership_id = '00000000-0000-0000-0000-000000000000' THEN NULL::uuid
           ELSE membership_id
           END AS membership_id,
       CASE
           WHEN drop_in_price = 9999 THEN NULL::numeric
           ELSE drop_in_price
           END AS payg_price,
       CASE
           WHEN program_price = 9999 THEN NULL::numeric
           ELSE program_price
           END AS program_price
FROM prepared_data;