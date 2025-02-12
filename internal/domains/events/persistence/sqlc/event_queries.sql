-- name: CreateEvent :execrows
INSERT INTO events (begin_time, end_time, facility_id, course_id, day)
VALUES ($1, $2, $3, $4, $5);

-- name: GetEvents :many
SELECT e.id, 
       begin_time, 
       end_time, 
       e.day, 
       c.id as course_id,
       c.name as course, 
         f.id as facility_id,
       f.name as facility
FROM events e
JOIN courses c ON c.id = e.course_id
JOIN facilities f ON f.id = e.facility_id
WHERE 
    (begin_time >= $1 OR $1::text LIKE '%00:00:00%')
    AND (end_time <= $2 OR $2::text LIKE '%00:00:00%')
    AND (facility_id = $3 OR $3 = '00000000-0000-0000-0000-000000000000')
    AND (course_id = $4 or $4 IS NULL);

-- name: GetEventById :one
SELECT e.id, begin_time, end_time, e.day, c.id as course_id, c.name as course, f.id as facility_id, f.name as facility
FROM events e
JOIN courses c ON c.id = e.course_id
JOIN facilities f ON f.id = e.facility_id
WHERE e.id = $1;

-- name: UpdateEvent :one
WITH updated AS (
    UPDATE events e
    SET begin_time = $1, end_time = $2, facility_id = $3, course_id = $4, day = $5
    WHERE e.id = $6
    RETURNING e.*
)
SELECT u.*, c.name as course_name, f.name as facility_name
FROM updated u
JOIN courses c ON c.id = u.course_id
JOIN facilities f ON f.id = u.facility_id;

-- name: DeleteEvent :execrows
DELETE FROM events WHERE id = $1;