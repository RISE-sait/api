-- name: CreateEvent :exec
INSERT INTO events.events (program_start_at, program_end_at, event_start_time, event_end_time, day, location_id,
                           program_id, capacity)
VALUES ($1, $2, $3, $4, $5,
        $6, $7, $8);

-- name: GetEvents :many
SELECT e.*,
       p.name        as program_name,
       p.description as program_description,
       p."type"      as program_type,
       l.name        as location_name,
       l.address     as address
FROM events.events e
         LEFT JOIN program.programs p ON e.program_id = p.id
         LEFT JOIN location.locations l ON e.location_id = l.id
WHERE (sqlc.narg('program_id') = program_id OR sqlc.narg('program_id') IS NULL)
  AND (sqlc.narg('location_id') = location_id OR sqlc.narg('location_id') IS NULL)
  AND (sqlc.narg('before') >= e.program_start_at OR sqlc.narg('before') IS NULL) -- within boundary
  AND (sqlc.narg('after') <= e.program_end_at OR sqlc.narg('after') IS NULL);

-- name: GetEventById :one
SELECT e.*,
       p.name        as program_name,
       p.description as program_description,
       p."type"      as program_type,
       l.name        as location_name,
       l.address     as address
FROM events.events e
          LEFT JOIN program.programs p ON e.program_id = p.id
         LEFT JOIN location.locations l ON e.location_id = l.id
WHERE e.id = $1;

-- name: UpdateEvent :exec
UPDATE events.events
SET program_start_at   = $1,
    program_end_at     = $2,
    location_id    = $3,
    program_id    = $4,
    event_start_time = $5,
    event_end_time   = $6,
    day                = $7,
    capacity = $8,
    updated_at     = current_timestamp
WHERE id = $9;

-- name: DeleteEvent :exec
DELETE
FROM events.events
WHERE id = $1;