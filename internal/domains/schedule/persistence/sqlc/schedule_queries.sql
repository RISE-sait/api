-- name: CreateSchedule :execrows
INSERT INTO schedules (begin_time, end_time, facility_id, course_id, day)
VALUES ($1, $2, $3, $4, $5);

-- name: GetSchedules :many
SELECT s.id, begin_time, end_time, s.day, c.name as course, f.name as facility FROM schedules s
JOIN courses c ON c.id = s.course_id
JOIN facilities f ON f.id = s.facility_id
WHERE 
    (begin_time >= $1 OR $1::text LIKE '%00:00:00%')
    AND (end_time <= $2 OR $2::text LIKE '%00:00:00%')
   AND (facility_id = $3 OR $3 = '00000000-0000-0000-0000-000000000000')
    AND (course_id = $4 or $4 IS NULL);

-- name: UpdateSchedule :execrows
UPDATE schedules s
SET begin_time = $1, end_time = $2, facility_id = $3, course_id = $4, day = $5
WHERE s.id = $6;

-- name: DeleteSchedule :execrows
DELETE FROM schedules WHERE id = $1;