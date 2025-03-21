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

-- name: InsertEvents :many
WITH events_data AS (SELECT unnest(@program_start_at_array::timestamptz[]) as program_start_at,
                            unnest(@program_end_at_array::timestamptz[])   as program_end_at,
                            unnest(@event_start_time_array::timetz[]) AS event_start_time,
                            unnest(@event_end_time_array::timetz[])   AS event_end_time,
                            unnest(@day_array::day_enum[])                 AS day,
                            unnest(@practice_name_array::text[])           AS practice_name,
                            unnest(@course_name_array::text[])             AS course_name,
                            unnest(@game_name_array::text[])               AS game_name,
                            unnest(@location_name_array::text[])           as location_name)
INSERT
INTO public.events (program_start_at, program_end_at, event_start_time, event_end_time, day, practice_id, course_id,
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
         LEFT JOIN LATERAL (SELECT id FROM course.courses WHERE name = e.course_name) c ON TRUE
         LEFT JOIN LATERAL (SELECT id FROM public.games WHERE name = e.game_name) g ON TRUE
         LEFT JOIN LATERAL (SELECT id FROM location.locations WHERE name = e.location_name) l ON TRUE
RETURNING id;

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

-- name: InsertBarberEvents :exec
WITH prepared_data AS (SELECT unnest(@begin_date_time_array::timestamptz[]) AS begin_date_time,
                              unnest(@end_date_time_array::timestamptz[])   AS end_date_time,
                              unnest(@customer_id_array::uuid[])            AS customer_id,
                              unnest(@barber_email_array::text[])           AS barber_email),
     user_data AS (SELECT pd.begin_date_time,
                          pd.end_date_time,
                          pd.customer_id,
                          ub.id AS barber_id
                   FROM prepared_data pd
                            LEFT JOIN
                        users.users ub ON pd.barber_email = ub.email)
INSERT
INTO barber.barber_events (begin_date_time, end_date_time, customer_id, barber_id)
SELECT begin_date_time,
       end_date_time,
       customer_id,
       barber_id
FROM user_data
WHERE customer_id IS NOT NULL
  AND barber_id IS NOT NULL
ON CONFLICT DO NOTHING;