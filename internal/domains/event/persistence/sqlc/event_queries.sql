-- name: CreateEvent :one
INSERT INTO events (event_start_at, event_end_at, session_start_time, session_end_time, day, location_id, course_id, practice_id, game_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetEvents :many
SELECT *
FROM events WHERE
(course_id = $1 OR $1 IS NULL)
    AND (practice_id = $2 or $2 IS NULL)
        AND (game_id = $3 or $3 IS NULL)
AND (location_id = $4 or $4 IS NULL);

-- name: GetEventById :one
SELECT *
FROM events
WHERE id = $1;

-- name: UpdateEvent :one
UPDATE events
SET event_start_at = $1, event_end_at = $2, session_start_time = $3, session_end_time = $4,
    location_id = $5, practice_id = $6, course_id = $7, game_id = $8, updated_at = current_timestamp
WHERE id = $9
RETURNING *;


-- name: DeleteEvent :execrows
DELETE FROM events WHERE id = $1;