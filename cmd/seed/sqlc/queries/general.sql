-- name: InsertLocations :exec
INSERT INTO location.locations (name, address)
VALUES (unnest(@name_array::text[]), unnest(@address_array::text[]))
RETURNING id;

-- name: InsertPractices :exec
INSERT INTO public.practices (name, description, level, capacity)
VALUES (unnest(@name_array::text[]),
        unnest(@description_array::text[]),
        unnest(@level_array::practice_level[]),
        unnest(@capacity_array::int[]))
RETURNING id;

-- name: InsertCourses :exec
INSERT INTO course.courses (name, description, capacity)
VALUES (unnest(@name_array::text[]),
        unnest(@description_array::text[]),
        unnest(@capacity_array::int[]))
RETURNING id;

-- name: InsertGames :exec
INSERT INTO public.games (name)
VALUES (unnest(@name_array::text[]))
RETURNING id;

-- name: InsertEvents :exec
WITH events_data AS (SELECT unnest(@program_start_at_array::timestamptz[]) as program_start_at,
                            unnest(@program_end_at_array::timestamptz[])   as program_end_at,
                            unnest(@session_start_time_array::timetz[])    AS session_start_time,
                            unnest(@session_end_time_array::timetz[])      AS session_end_time,
                            unnest(@day_array::day_enum[])                 AS day,
                            unnest(@practice_name_array::text[])           AS practice_name,
                            unnest(@course_name_array::text[])             AS course_name,
                            unnest(@game_name_array::text[])               AS game_name,
                            unnest(@location_name_array::text[])           as location_name)
INSERT
INTO public.events (program_start_at, program_end_at, session_start_time, session_end_time, day, practice_id, course_id,
                    game_id, location_id)
SELECT e.program_start_at,
       e.program_end_at,
       e.session_start_time,
       e.session_end_time,
       e.day,
       p.id AS practice_id,
       c.id AS course_id,
       g.id AS game_id,
       l.id AS location_id
FROM events_data e
         LEFT JOIN LATERAL (SELECT id FROM public.practices WHERE name = e.practice_name) p ON TRUE
         LEFT JOIN LATERAL (SELECT id FROM course.courses WHERE name = e.course_name) c ON TRUE
         LEFT JOIN LATERAL (SELECT id FROM public.games WHERE name = e.game_name) g ON TRUE
         LEFT JOIN LATERAL (SELECT id FROM location.locations WHERE name = e.location_name) l ON TRUE;

-- name: InsertCustomersEnrollments :many
WITH prepared_data AS (SELECT unnest(@customer_id_array::uuid[])          AS customer_id,
                              unnest(@event_id_array::uuid[])             AS event_id,
                              unnest(@checked_in_at_array::timestamptz[]) AS raw_checked_in_at,
                              unnest(@is_cancelled_array::bool[])         AS is_cancelled)
INSERT
INTO public.customer_enrollment(customer_id, event_id, checked_in_at, is_cancelled)
SELECT customer_id,
       event_id,
       NULLIF(raw_checked_in_at, '0001-01-01 00:00:00 UTC') AS checked_in_at,
       is_cancelled
FROM prepared_data
RETURNING id;