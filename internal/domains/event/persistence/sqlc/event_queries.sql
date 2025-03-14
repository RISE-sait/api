-- name: CreateEvent :one
INSERT INTO events (event_start_at, event_end_at, location_id, course_id, practice_id, game_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetEvents :many
SELECT *
FROM events
WHERE event_start_at >= sqlc.arg('after')
  AND event_end_at <= sqlc.arg('before')
  AND (sqlc.narg('course_id') = course_id OR sqlc.narg('course_id') IS NULL)
  AND (sqlc.narg('game_id') = game_id OR sqlc.narg('game_id') IS NULL)
  AND (sqlc.narg('practice_id') = practice_id OR sqlc.narg('practice_id') IS NULL)
  AND (sqlc.narg('location_id') = location_id OR sqlc.narg('location_id') IS NULL);

-- name: GetEventById :one
SELECT *
FROM events
WHERE id = $1;

-- name: UpdateEvent :one
UPDATE events
SET event_start_at = $1,
    event_end_at   = $2,
    location_id    = $3,
    practice_id    = $4,
    course_id      = $5,
    game_id        = $6,
    updated_at     = current_timestamp
WHERE id = $7
RETURNING *;


-- name: DeleteEvent :exec
DELETE
FROM events
WHERE id = $1;