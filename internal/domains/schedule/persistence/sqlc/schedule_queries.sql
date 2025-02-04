-- name: CreateSchedule :execrows
INSERT INTO schedules (begin_datetime, end_datetime, facility_id, course_id, day)
VALUES ($1, $2, $3, $4, $5);

-- name: GetSchedules :many
SELECT s.id, begin_datetime, end_datetime, s.day, c.name as course, f.name as facility FROM schedules s
JOIN courses c ON c.id = s.course_id
JOIN facilities f ON f.id = s.facility_id
WHERE 
    (begin_datetime >= $1 OR $1::text LIKE '0001-01-01%')
    AND (end_datetime <= $2 OR $2::text LIKE '0001-01-01%')
   AND (facility_id = $3 OR $3 = '00000000-0000-0000-0000-000000000000')
    AND (course_id = $4 or $4 IS NULL);

-- name: UpdateSchedule :execrows
UPDATE schedules s
SET begin_datetime = $1, end_datetime = $2, facility_id = $3, course_id = $4, day = $5
WHERE s.id = $6;

-- name: DeleteSchedule :execrows
DELETE FROM schedules WHERE id = $1;