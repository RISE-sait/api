-- name: InsertEvents :many
WITH events_data AS (SELECT unnest(@program_start_at_array::timestamptz[]) as program_start_at,
                            unnest(@program_end_at_array::timestamptz[])   as program_end_at,
                            unnest(@event_start_time_array::timetz[])      AS event_start_time,
                            unnest(@event_end_time_array::timetz[])        AS event_end_time,
                            unnest(@day_array::day_enum[])                 AS day,
                            unnest(@program_name_array::text[])           AS program_name,
                            unnest(@location_name_array::text[])           as location_name)
INSERT
INTO events.events (program_start_at, program_end_at, event_start_time, event_end_time, program_id, day, location_id)
SELECT e.program_start_at,
       e.program_end_at,
       e.event_start_time,
       e.event_end_time,
         p.id AS program_id,
       e.day,
       l.id AS location_id
FROM events_data e
         LEFT JOIN LATERAL (SELECT id FROM program.programs WHERE name = e.program_name) p ON TRUE
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