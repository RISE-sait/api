-- name: InsertEvents :many
WITH events_data AS (SELECT unnest(@program_start_at_array::timestamptz[]) as program_start_at,
                            unnest(@program_end_at_array::timestamptz[])   as program_end_at,
                            unnest(@event_start_time_array::timetz[])      AS event_start_time,
                            unnest(@event_end_time_array::timetz[])        AS event_end_time,
                            unnest(@day_array::day_enum[])                 AS day,
                            unnest(@practice_name_array::text[])           AS practice_name,
                            unnest(@course_name_array::text[])             AS course_name,
                            unnest(@game_name_array::text[])               AS game_name,
                            unnest(@location_name_array::text[])           as location_name)
INSERT
INTO events.events (program_start_at, program_end_at, event_start_time, event_end_time, day, practice_id, course_id,
                    game_id, location_id)
SELECT e.program_start_at,
       e.program_end_at,
       e.event_start_time,
       e.event_end_time,
       e.day,
       p.id AS practice_id,
       c.id AS course_id,
       g.id AS game_id,
       l.id AS location_id
FROM events_data e
         LEFT JOIN LATERAL (SELECT id FROM public.practices WHERE name = e.practice_name) p ON TRUE
         LEFT JOIN LATERAL (SELECT id FROM courses WHERE name = e.course_name) c ON TRUE
         LEFT JOIN LATERAL (SELECT id FROM public.games WHERE name = e.game_name) g ON TRUE
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