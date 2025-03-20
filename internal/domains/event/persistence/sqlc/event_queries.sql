-- name: CreateEvent :exec
INSERT INTO events (program_start_at, program_end_at, session_start_time, session_end_time, day, location_id, course_id,
                    practice_id, game_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: GetEvents :many
SELECT *
FROM events
WHERE (sqlc.narg('course_id') = course_id OR sqlc.narg('course_id') IS NULL)
  AND (sqlc.narg('game_id') = game_id OR sqlc.narg('game_id') IS NULL)
  AND (sqlc.narg('practice_id') = practice_id OR sqlc.narg('practice_id') IS NULL)
  AND (sqlc.narg('location_id') = location_id OR sqlc.narg('location_id') IS NULL)
  AND (sqlc.narg('before') >= events.program_start_at OR sqlc.narg('before') IS NULL) -- within boundary
  AND (sqlc.narg('after') <= events.program_end_at OR sqlc.narg('after') IS NULL);

-- name: GetEventById :one
SELECT *
FROM events
WHERE id = $1;

-- name: UpdateEvent :exec
UPDATE events
SET program_start_at   = $1,
    program_end_at     = $2,
    location_id    = $3,
    practice_id    = $4,
    course_id      = $5,
    game_id        = $6,
    session_start_time = $7,
    session_end_time   = $8,
    day                = $9,
    updated_at     = current_timestamp
WHERE id = $10;

-- name: DeleteEvent :exec
DELETE
FROM events
WHERE id = $1;