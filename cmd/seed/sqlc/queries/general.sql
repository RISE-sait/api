-- name: InsertLocations :exec
INSERT INTO location.locations (name, address)
VALUES (unnest(@name_array::text[]), unnest(@address_array::text[]))
RETURNING id;

-- name: InsertPractices :exec
WITH prepared_data as (
        SELECT unnest(@name_array::text[]) as name,
        unnest(@description_array::text[]) as description,
        unnest(@level_array::program.program_level[]) as level)
INSERT INTO program.programs (name, description, type, level)
SELECT name,
       description,
       'practice',
       level
FROM prepared_data;

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

-- name: InsertEvents :many
WITH events_data AS (SELECT unnest(@program_start_at_array::timestamptz[]) as program_start_at,
                            unnest(@program_end_at_array::timestamptz[])   as program_end_at,
                            unnest(@event_start_time_array::timetz[]) AS event_start_time,
                            unnest(@event_end_time_array::timetz[])   AS event_end_time,
                            unnest(@day_array::day_enum[])                 AS day,
                            unnest(@program_name_array::text[])           AS program_name,
                            unnest(@location_name_array::text[])           as location_name)
INSERT
INTO events.events (program_start_at, program_end_at, event_start_time, event_end_time, day, program_id, location_id)
SELECT e.program_start_at,
       e.program_end_at,
       e.event_start_time,
       e.event_end_time,
       e.day,
       p.id AS program_id,
       l.id AS location_id
FROM events_data e
         LEFT JOIN LATERAL (SELECT id FROM program.programs p WHERE p.name = e.program_name) p ON TRUE
         LEFT JOIN LATERAL (SELECT id FROM location.locations WHERE name = e.location_name) l ON TRUE
RETURNING id;

-- name: InsertCustomersEnrollments :many
WITH prepared_data AS (SELECT unnest(@customer_id_array::uuid[])          AS customer_id,
                              unnest(@event_id_array::uuid[])             AS event_id,
                              unnest(@checked_in_at_array::timestamptz[]) AS raw_checked_in_at,
                              unnest(@is_cancelled_array::bool[])         AS is_cancelled)
INSERT
INTO events.customer_enrollment(customer_id, event_id, checked_in_at, is_cancelled)
SELECT customer_id,
       event_id,
       NULLIF(raw_checked_in_at, '0001-01-01 00:00:00 UTC') AS checked_in_at,
       is_cancelled
FROM prepared_data
RETURNING id;


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