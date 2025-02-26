-- name: CreateEvent :one
INSERT INTO events (begin_time, end_time, location_id, course_id, practice_id, day)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetEvents :many
SELECT *
FROM events WHERE
course_id = $1 OR $1 IS NULL
    AND (practice_id = $2 or $2 IS NULL)
AND location_id = sqlc.narg('location_id') or sqlc.narg('location_id') IS NULL;

-- name: GetEventById :one
SELECT *
FROM events
WHERE id = $1;

-- name: UpdateEvent :one
UPDATE events
    SET begin_time = $1, end_time = $2, location_id = $3, practice_id = $4, day = $5, course_id = $6
    WHERE id = $7
    RETURNING *;

-- name: DeleteEvent :execrows
DELETE FROM events WHERE id = $1;