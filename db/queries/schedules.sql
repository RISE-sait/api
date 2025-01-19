-- Get all schedules
-- name: GetAllSchedules :many
SELECT * FROM schedules;

-- Get schedule by course_id
-- name: GetScheduleByCourseID :one
SELECT * FROM schedules WHERE course_id = $1;

-- Create a new schedule
-- name: CreateSchedule :execrows
INSERT INTO schedules (begin_datetime, end_datetime, course_id, facility_id, day)
VALUES ($1, $2, $3, $4, $5);

-- Update a schedule
-- name: UpdateSchedule :execrows
UPDATE schedules
SET begin_datetime = $1, end_datetime = $2, course_id = $3, facility_id = $4, day = $5
WHERE course_id = $6;

-- Delete a schedule
-- name: DeleteSchedule :execrows
DELETE FROM schedules WHERE course_id = $1;