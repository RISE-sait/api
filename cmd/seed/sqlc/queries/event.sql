-- name: InsertEvents :many
WITH events_data AS (SELECT unnest(@start_at_array::timestamptz[])     as start_at,
                            unnest(@end_at_array::timestamptz[])       as end_at,
                            unnest(@location_name_array::varchar[])    as location_name,
                            unnest(@created_by_email_array::varchar[]) AS created_by_email,
                            unnest(@updated_by_email_array::varchar[]) AS updated_by_email,
                            unnest(@capacity_array::int[])             AS capacity,
                            unnest(@schedule_id_array::uuid[])         AS schedule_id)
INSERT
INTO events.events (start_at, end_at, location_id, created_by, updated_by, capacity, schedule_id)
SELECT e.start_at,
       e.end_at,
       l.id,
       creator.id,
       updater.id,
       e.capacity,
       NULLIF(e.schedule_id, '00000000-0000-0000-0000-000000000000')
FROM events_data e
         LEFT JOIN users.users creator ON creator.email = e.created_by_email
         LEFT JOIN users.users updater ON updater.email = e.updated_by_email
         JOIN location.locations l ON l.name = e.location_name
RETURNING id;


-- name: InsertSchedule :one
INSERT INTO public.schedules (recurrence_start_at,
                              recurrence_end_at,
                              event_start_time,
                              event_end_time,
                              program_id,
                              day,
                              location_id)
VALUES ($1,
        $2,
        $3,
        $4,
        (SELECT id FROM program.programs p WHERE p.name = sqlc.narg('program_name')),
        $5,
        (SELECT id FROM location.locations l WHERE l.name = sqlc.arg('location_name')))
RETURNING id;


-- name: InsertCustomersEnrollments :many
WITH prepared_data AS (SELECT unnest(@customer_id_array::uuid[])          AS customer_id,
                              unnest(@event_id_array::uuid[])             AS event_id,
                              unnest(@checked_in_at_array::timestamptz[]) AS raw_checked_in_at)
INSERT
INTO events.customer_enrollment(customer_id, event_id, checked_in_at)
SELECT customer_id,
       event_id,
       NULLIF(raw_checked_in_at, '0001-01-01 00:00:00 UTC') AS checked_in_at
FROM prepared_data
RETURNING id;

-- name: InsertEventsStaff :exec
WITH prepared_data AS (SELECT unnest(@event_id_array::uuid[]) AS event_id,
                              unnest(@staff_id_array::uuid[]) AS staff_id)
INSERT
INTO events.staff(event_id, staff_id)
SELECT event_id,
       staff_id
FROM prepared_data;