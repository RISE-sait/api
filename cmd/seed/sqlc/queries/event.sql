-- name: InsertEvents :many
WITH events_data AS (
    SELECT 
        unnest(@start_at_array::timestamptz[])        AS start_at,
        unnest(@end_at_array::timestamptz[])          AS end_at,
        unnest(@program_name_array::varchar[])        AS program_name,
        unnest(@location_name_array::varchar[])       AS location_name,
        unnest(@created_by_email_array::varchar[])    AS created_by_email,
        unnest(@updated_by_email_array::varchar[])    AS updated_by_email
)
INSERT INTO events.events (
    start_at, end_at, program_id, location_id, created_by, updated_by
)
SELECT 
    e.start_at,
    e.end_at,
    p.id,
    l.id,
    creator.id,
    updater.id
FROM events_data e
JOIN program.programs p 
  ON p.type = LOWER(e.program_name)::program.program_type
JOIN users.users creator 
  ON creator.email = e.created_by_email
JOIN users.users updater 
  ON updater.email = e.updated_by_email
JOIN location.locations l 
  ON l.name = e.location_name
ON CONFLICT DO NOTHING
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