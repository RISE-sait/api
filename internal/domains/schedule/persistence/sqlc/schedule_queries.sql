-- name: CreateSchedule :execrows
INSERT INTO schedules (begin_datetime, end_datetime, facility_id, course_id, day)
VALUES ($1, $2, $3, $4, $5);

-- name: GetSchedules :many
SELECT * FROM schedules
WHERE 
    (begin_datetime >= $1 OR $1 = '0001-01-01')
    AND (end_datetime <= $2 OR $2 = '0001-01-01')
   AND (facility_id = $3 OR $3 = '00000000-0000-0000-0000-000000000000')
    AND (course_id = $4 or $4 = '00000000-0000-0000-0000-000000000000');

-- name: UpdateSchedule :execrows
UPDATE schedules s
SET begin_datetime = $1, end_datetime = $2, facility_id = $3, course_id = $4, day = $5
WHERE s.id = $6;

-- name: DeleteSchedule :execrows
DELETE FROM schedules WHERE id = $1;