-- name: CreateEvent :one
INSERT INTO events (begin_date_time, end_date_time, location_id, course_id, practice_id, game_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetEvents :many
SELECT *
FROM events WHERE
course_id = $1 OR $1 IS NULL
    AND (practice_id = $2 or $2 IS NULL)
        AND (game_id = $3 or $3 IS NULL)

AND location_id = sqlc.narg('location_id') or sqlc.narg('location_id') IS NULL;

-- name: GetEventById :one
SELECT *
FROM events
WHERE id = $1;

-- name: UpdateEvent :one
UPDATE events
    SET begin_date_time = $1, end_date_time = $2, location_id = $3, practice_id = $4, course_id = $5,
        game_id = $6
    WHERE id = $7
    RETURNING *;

-- name: DeleteEvent :execrows
DELETE FROM events WHERE id = $1;