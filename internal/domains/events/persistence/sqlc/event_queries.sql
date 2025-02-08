-- name: CreateEvent :execrows
INSERT INTO events (begin_time, end_time, facility_id, course_id, day)
VALUES ($1, $2, $3, $4, $5);

-- name: GetEvents :many
SELECT e.id, begin_time, end_time, e.day, c.name as course, f.name as facility FROM events e
JOIN courses c ON c.id = e.course_id
JOIN facilities f ON f.id = e.facility_id
WHERE 
    (begin_time >= $1 OR $1::text LIKE '%00:00:00%')
    AND (end_time <= $2 OR $2::text LIKE '%00:00:00%')
    AND (facility_id = $3 OR $3 = '00000000-0000-0000-0000-000000000000')
    AND (course_id = $4 or $4 IS NULL);

-- name: UpdateEvent :execrows
UPDATE events e
SET begin_time = $1, end_time = $2, facility_id = $3, course_id = $4, day = $5
WHERE e.id = $6;

-- name: DeleteEvent :execrows
DELETE FROM events WHERE id = $1;

-- name: GetCustomersCountByEventId :one
SELECT COUNT(id) from customer_events WHERE event_id = $1;
