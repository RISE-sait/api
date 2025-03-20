-- name: CreateEvent :exec
INSERT INTO events (program_start_at, program_end_at, event_start_time, event_end_time, day, location_id, course_id,
                    practice_id, game_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: GetEvents :many
SELECT e.*,
       p.name        as practice_name,
       p.description as practice_description,
       c.name        as course_name,
       c.description as course_description,
       g.name        as game_name,
       l.name        as location_name,
       l.address     as address
FROM public.events e
         LEFT JOIN public.practices p ON e.practice_id = p.id
         LEFT JOIN course.courses c ON e.course_id = c.id
         LEFT JOIN public.games g ON e.game_id = g.id
         LEFT JOIN location.locations l ON e.location_id = l.id
WHERE (sqlc.narg('course_id') = course_id OR sqlc.narg('course_id') IS NULL)
  AND (sqlc.narg('game_id') = game_id OR sqlc.narg('game_id') IS NULL)
  AND (sqlc.narg('practice_id') = practice_id OR sqlc.narg('practice_id') IS NULL)
  AND (sqlc.narg('location_id') = location_id OR sqlc.narg('location_id') IS NULL)
  AND (sqlc.narg('before') >= e.program_start_at OR sqlc.narg('before') IS NULL) -- within boundary
  AND (sqlc.narg('after') <= e.program_end_at OR sqlc.narg('after') IS NULL);

-- name: GetEventById :one
SELECT e.*,
       p.name        as practice_name,
       p.description as practice_description,
       c.name        as course_name,
       c.description as course_description,
       g.name        as game_name,
       l.name        as location_name,
       l.address     as address
FROM public.events e
         LEFT JOIN public.practices p ON e.practice_id = p.id
         LEFT JOIN course.courses c ON e.course_id = c.id
         LEFT JOIN public.games g ON e.game_id = g.id
         LEFT JOIN location.locations l ON e.location_id = l.id
WHERE e.id = $1;

-- name: UpdateEvent :exec
UPDATE events
SET program_start_at   = $1,
    program_end_at     = $2,
    location_id    = $3,
    practice_id    = $4,
    course_id      = $5,
    game_id        = $6,
    event_start_time = $7,
    event_end_time   = $8,
    day                = $9,
    updated_at     = current_timestamp
WHERE id = $10;

-- name: DeleteEvent :exec
DELETE
FROM events
WHERE id = $1;